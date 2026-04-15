package mpr

import "go.mongodb.org/mongo-driver/bson"

const microflowDocumentType = "Microflows$Microflow"

func addMicroflowPseudocode(name string, attributes bson.M) {
	cleanedData := bsonToMap(attributes)
	pseudocode, err := generateMicroflowPseudocode(name, cleanedData)
	if err != nil {
		log.Warnf("Could not generate pseudocode for microflow %s: %v", name, err)
		return
	}
	attributes["pseudocode"] = pseudocode
}
