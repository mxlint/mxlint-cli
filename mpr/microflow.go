package mpr

func transformMicroflow(mf MxDocument) MxDocument {
	// Transform a microflow
	log.Infof("Transforming microflow %s", mf.Name)
	if mf.Attributes["Name"] != "MicroflowSimple" {
		return mf
	}

	cleanedData := bsonToMap(mf.Attributes)
	objsCollection := cleanedData["ObjectCollection"].(map[string]interface{})
	objs := objsCollection["Objects"].([]map[string]interface{})
	log.Infof("Objects: %v", objs)
	flows := cleanedData["Flows"].([]map[string]interface{})

	startEvent := getMxMicroflowObjectByType(objs, "Microflows$StartEvent")
	endEvent := getMxMicroflowObjectByType(objs, "Microflows$EndEvent")
	log.Infof("StartEvent: %v", startEvent)
	log.Infof("EndEvent: %v", endEvent)

	tree := buildSequence(startEvent, flows, objs)
	log.Infof("Tree: %v", tree)
	mf.Attributes["Sequence"] = tree

	return mf
}

func buildSequence(object MxMicroflowObject, flows []map[string]interface{}, objects []map[string]interface{}) []MxMicroflowObject {

	current := make([]MxMicroflowObject, 0)
	current = append(current, object)
	next := getMxMicroflowFlow(flows, objects, object.ID)
	for next.ID != "" {
		current = append(current, next)
		next = getMxMicroflowFlow(flows, objects, next.ID)
	}
	return current
}

func getMxMicroflowFlow(flows []map[string]interface{}, objects []map[string]interface{}, originId string) MxMicroflowObject {
	// Get a microflow object
	for _, flow := range flows {
		if flow["OriginPointer"].(string) == originId {
			destinationId := flow["DestinationPointer"].(string)
			return getMxMicroflowObjectByID(objects, destinationId)
		}
	}

	return MxMicroflowObject{}
}

func getMxMicroflowObjectByType(objs []map[string]interface{}, objType string) MxMicroflowObject {
	// Get a microflow object
	for _, obj := range objs {
		if obj["$Type"] == objType {
			return MxMicroflowObject{
				Type:       obj["$Type"].(string),
				ID:         obj["$ID"].(string),
				Attributes: obj,
			}
		}
	}

	return MxMicroflowObject{}
}

func getMxMicroflowObjectByID(objs []map[string]interface{}, objID string) MxMicroflowObject {
	// Get a microflow object
	for _, obj := range objs {
		if obj["$ID"] == objID {
			return MxMicroflowObject{
				Type:       obj["$Type"].(string),
				ID:         obj["$ID"].(string),
				Attributes: obj,
			}
		}
	}

	return MxMicroflowObject{}
}
