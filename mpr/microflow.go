package mpr

func transformMicroflow(mf MxDocument) MxDocument {
	// Transform a microflow
	log.Infof("Transforming microflow %s", mf.Name)
	if !Contains([]string{"MicroflowSimple", "MicroflowSplit", "MicroflowSplitThenMerge", "MicroflowComplexSplit"}, mf.Name) {
		return mf
	}

	cleanedData := bsonToMap(mf.Attributes)
	objsCollection := cleanedData["ObjectCollection"].(map[string]interface{})
	objs := convertToMxMicroflowObjects(objsCollection["Objects"].([]interface{}))
	flows := convertToMxMicroflowEdges(cleanedData["Flows"].([]interface{}))

	startEvent := getMxMicroflowObjectByType(objs, "Microflows$StartEvent")

	root := MxMicroflowNode{
		Type:       startEvent.Type,
		ID:         startEvent.ID,
		Attributes: startEvent.Attributes,
	}
	buildDAG(&root, nil, flows, objs)

	mainFlow := make([]map[string]interface{}, 0)
	extractMainFlow(&mainFlow, &root)
	mf.Attributes["MainFunction"] = mainFlow
	for _, obj := range mainFlow {
		log.Infof("XXX: %v", obj)
	}

	return mf
}

func extractMainFlow(mainFlow *[]map[string]interface{}, current *MxMicroflowNode) {
	c := convertMxMicroflowNodeToMap(current)
	log.Infof("Current: %v", c)
	*mainFlow = append(*mainFlow, c)
	if current.Type == "Microflows$EndEvent" {
		return
	}

	children := current.Children
	current.Children = nil
	if children == nil {
		return
	}
	if len(*children) == 0 {
		return
	} else if len(*children) == 1 {
		// sequence
		child := (*children)[0]
		extractMainFlow(mainFlow, &child)
		return
	} else {
		// split
		splits := make([]interface{}, 0)
		for _, child := range *children {
			subflow := make([]map[string]interface{}, 0)
			log.Infof("Split: %v", child)
			extractMainFlow(&subflow, &child)
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
		// FIXME: Check if there is a loop
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
	case "Microflows$SequenceFlow":
		obj := getMxMicroflowObjectByID(objects, current.Attributes["DestinationPointer"].(string))
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

func getMxMicroflowEdgesByOrigin(edges []MxMicroflowEdge, originId string) []MxMicroflowEdge {
	result := make([]MxMicroflowEdge, 0)
	for _, edge := range edges {
		if edge.Origin == originId {
			result = append(result, edge)
		}
	}
	return result
}

func getMxMicroflowObjectByType(objs []MxMicroflowObject, objType string) MxMicroflowObject {
	for _, obj := range objs {
		if obj.Type == objType {
			return obj
		}
	}
	panic("MPR file probably corrupted")
}

func getMxMicroflowObjectByID(objs []MxMicroflowObject, objID string) MxMicroflowObject {
	for _, obj := range objs {
		if obj.ID == objID {
			return obj
		}
	}
	panic("MPR file probably corrupted")
}

func convertToMxMicroflowObjects(objs []interface{}) []MxMicroflowObject {
	result := make([]MxMicroflowObject, len(objs))
	for _, o := range objs {
		castedObject := o.(map[string]interface{})
		result = append(result, MxMicroflowObject{
			Type:       castedObject["$Type"].(string),
			ID:         castedObject["$ID"].(string),
			Attributes: castedObject,
		})
	}
	return result
}

func convertToMxMicroflowEdges(flows []interface{}) []MxMicroflowEdge {
	result := make([]MxMicroflowEdge, len(flows))
	for _, f := range flows {
		castedFlow := f.(map[string]interface{})
		result = append(result, MxMicroflowEdge{
			Type:        castedFlow["$Type"].(string),
			ID:          castedFlow["$ID"].(string),
			Origin:      castedFlow["OriginPointer"].(string),
			Destination: castedFlow["DestinationPointer"].(string),
			Attributes:  castedFlow,
		})
	}
	return result
}
