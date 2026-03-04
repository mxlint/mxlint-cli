package mpr

const microflowDocumentType = "Microflows$Microflow"

func enrichMicroflowDocument(mf MxDocument, mode string) MxDocument {
	mf = addMicroflowPseudocode(mf)
	if mode == "advanced" {
		mf = transformMicroflow(mf)
	}
	return mf
}

func transformMicroflow(mf MxDocument) MxDocument {
	log.Infof("Transforming microflow %s", mf.Name)

	cleanedData := bsonToMap(mf.Attributes)
	objs, flows, ok := extractMicroflowGraphData(mf.Name, cleanedData)
	if !ok {
		return mf
	}
	startEvent, ok := getMxMicroflowObjectByType(objs, "Microflows$StartEvent")
	if !ok {
		log.Warnf("StartEvent not found for microflow %s, skipping transformation", mf.Name)
		return mf
	}

	root := MxMicroflowNode{
		Type:       startEvent.Type,
		ID:         startEvent.ID,
		Attributes: startEvent.Attributes,
	}
	buildDAG(&root, nil, flows, objs)
	mainFlow := make([]map[string]interface{}, 0)
	labels := make(map[string]interface{}, 0)
	extractMainFlow(&mainFlow, &root, &labels)
	mf.Attributes["MainFunction"] = mainFlow
	// remove ObjectCollection
	delete(mf.Attributes, "ObjectCollection")
	return mf
}

func addMicroflowPseudocode(mf MxDocument) MxDocument {
	cleanedData := bsonToMap(mf.Attributes)
	pseudocode, err := generateMicroflowPseudocode(mf.Name, cleanedData)
	if err != nil {
		log.Warnf("Could not generate pseudocode for microflow %s: %v", mf.Name, err)
		return mf
	}
	mf.Attributes["pseudocode"] = pseudocode
	return mf
}

func extractMicroflowGraphData(name string, cleanedData map[string]interface{}) ([]MxMicroflowObject, []MxMicroflowEdge, bool) {
	objsCollectionRaw, ok := cleanedData["ObjectCollection"]
	if !ok || objsCollectionRaw == nil {
		log.Warnf("ObjectCollection not found for microflow %s, skipping transformation", name)
		return nil, nil, false
	}

	objsCollection, ok := objsCollectionRaw.(map[string]interface{})
	if !ok {
		log.Warnf("ObjectCollection is not a map for microflow %s, skipping transformation", name)
		return nil, nil, false
	}

	objectsRaw, ok := objsCollection["Objects"]
	if !ok || objectsRaw == nil {
		log.Warnf("Objects not found in ObjectCollection for microflow %s, skipping transformation", name)
		return nil, nil, false
	}

	objects, ok := objectsRaw.([]interface{})
	if !ok {
		log.Warnf("Objects is not a slice for microflow %s, skipping transformation", name)
		return nil, nil, false
	}

	flowsRaw, ok := cleanedData["Flows"]
	if !ok || flowsRaw == nil {
		log.Warnf("Flows not found for microflow %s, skipping transformation", name)
		return nil, nil, false
	}

	flowsSlice, ok := flowsRaw.([]interface{})
	if !ok {
		log.Warnf("Flows is not a slice for microflow %s, skipping transformation", name)
		return nil, nil, false
	}

	return convertToMxMicroflowObjects(objects), convertToMxMicroflowEdges(flowsSlice), true
}

func extractMainFlow(mainFlow *[]map[string]interface{}, current *MxMicroflowNode, labels *map[string]interface{}) {
	c := convertMxMicroflowNodeToMap(current)
	*mainFlow = append(*mainFlow, c)
	if current.Type == "Microflows$EndEvent" {
		return
	}

	if current.Type == "Microflows$ExclusiveMerge" {
		id, ok := c["ID"].(string)
		if !ok {
			log.Warn("ID is not a string or is nil")
			return
		}
		// check if label already exists
		if _, ok := (*labels)[id]; !ok {
			(*labels)[id] = c
			// continue expanding this subflow
		} else {
			log.Infof("Loop detected; not traversing")
			return
		}
	}

	children := current.Children
	//current.Children = nil
	if children == nil {
		return
	}
	if len(*children) == 0 {
		return
	} else if len(*children) == 1 {
		// sequence
		child := (*children)[0]
		extractMainFlow(mainFlow, &child, labels)
	} else {
		// split
		splits := make([]interface{}, 0)
		for _, child := range *children {
			subflow := make([]map[string]interface{}, 0)
			extractMainFlow(&subflow, &child, labels)
			splits = append(splits, subflow)
		}
		c["Splits"] = splits
	}
}

func buildDAG(current *MxMicroflowNode, parent *MxMicroflowNode, flows []MxMicroflowEdge, objects []MxMicroflowObject) {

	current.Parent = parent
	children := make([]MxMicroflowNode, 0)

	switch current.Type {
	case "Microflows$ExclusiveMerge":
		start := backtrack(current, current.Parent)
		if start == nil {
			// no loop
			edges := getMxMicroflowEdgesByOrigin(flows, current.ID)

			for _, edge := range edges {
				edgeNode := MxMicroflowNode{
					Type:       edge.Type,
					ID:         edge.ID,
					Attributes: edge.Attributes,
				}

				buildDAG(&edgeNode, current, flows, objects)
				children = append(children, edgeNode)
			}
		} else {
			log.Infof("Loop detected; not traversing")
			return
		}
	case "Microflows$SequenceFlow":
		destination, _ := current.Attributes["DestinationPointer"].(string)
		obj, ok := getMxMicroflowObjectByID(objects, destination)
		if !ok {
			log.Warnf("Destination object %s not found for sequence flow %s", destination, current.ID)
			return
		}
		objectNode := MxMicroflowNode{
			Type:       obj.Type,
			ID:         obj.ID,
			Attributes: obj.Attributes,
		}
		buildDAG(&objectNode, current, flows, objects)
		children = append(children, objectNode)
	default:
		edges := getMxMicroflowEdgesByOrigin(flows, current.ID)

		for _, edge := range edges {
			edgeNode := MxMicroflowNode{
				Type:       edge.Type,
				ID:         edge.ID,
				Attributes: edge.Attributes,
			}

			buildDAG(&edgeNode, current, flows, objects)
			children = append(children, edgeNode)
		}
	}
	current.Children = &children
}

func backtrack(current *MxMicroflowNode, parent *MxMicroflowNode) *MxMicroflowNode {
	if parent == nil {
		return nil
	}
	if parent.ID == current.ID {
		return parent
	}
	return backtrack(current, parent.Parent)
}

func getMxMicroflowEdgesByOrigin(edges []MxMicroflowEdge, originId string) []MxMicroflowEdge {
	result := make([]MxMicroflowEdge, 0)
	for _, edge := range edges {
		if edge.Origin == originId {
			result = append(result, edge)
		}
	}
	return result
}

func getMxMicroflowObjectByType(objs []MxMicroflowObject, objType string) (MxMicroflowObject, bool) {
	for _, obj := range objs {
		if obj.Type == objType {
			return obj, true
		}
	}
	return MxMicroflowObject{}, false
}

func getMxMicroflowObjectByID(objs []MxMicroflowObject, objID string) (MxMicroflowObject, bool) {
	for _, obj := range objs {
		if obj.ID == objID {
			return obj, true
		}
	}
	return MxMicroflowObject{}, false
}

func convertToMxMicroflowObjects(objs []interface{}) []MxMicroflowObject {
	result := make([]MxMicroflowObject, 0, len(objs))
	for _, o := range objs {
		castedObject, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		objType, _ := castedObject["$Type"].(string)
		id := readMicroflowID(castedObject["$ID"])
		if objType == "" || id == "" {
			continue
		}
		result = append(result, MxMicroflowObject{
			Type:       objType,
			ID:         id,
			Attributes: castedObject,
		})
	}
	return result
}

func convertToMxMicroflowEdges(flows []interface{}) []MxMicroflowEdge {
	result := make([]MxMicroflowEdge, 0, len(flows))
	for _, f := range flows {
		castedFlow, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		flowType, _ := castedFlow["$Type"].(string)
		id := readMicroflowID(castedFlow["$ID"])
		origin := readMicroflowID(castedFlow["OriginPointer"])
		destination := readMicroflowID(castedFlow["DestinationPointer"])
		if flowType == "" || id == "" || origin == "" || destination == "" {
			continue
		}
		result = append(result, MxMicroflowEdge{
			Type:        flowType,
			ID:          id,
			Origin:      origin,
			Destination: destination,
			Attributes:  castedFlow,
		})
	}
	return result
}
