package mpr

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var log = logrus.New() // Initialize with a default logger

// SetLogger allows the main application to set the logger, including its configuration.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

var ignoredAttributes = []string{"ID", "$ID", "Flows", "OriginPointer", "Type", "LineType", "DestinationPointer", "Image", "ImageData", "GUID", "StableId", "Size", "RelativeMiddlePoint", "Location", "OriginBezierVector", "DestinationBezierVector", "OriginConnectionIndex", "DestinationConnectionIndex"}

func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func ignoreAttributes(data bson.M, ignore []string) bson.M {
	result := make(bson.M)

	for key, value := range data {
		ignoreKey := false

		for _, ignoreAttr := range ignore {
			if key == ignoreAttr {
				ignoreKey = true
				break
			}
		}

		if !ignoreKey {
			if reflect.TypeOf(value) == reflect.TypeOf(primitive.A{}) {
				castedData := value.(primitive.A)
				var interfaceSlice []interface{} = castedData
				if len(interfaceSlice) > 0 {
					if reflect.TypeOf(interfaceSlice[0]) == reflect.TypeOf(int32(1)) {
						value = interfaceSlice[1:]
					} else {
						value = interfaceSlice
					}
				} else {
					value = interfaceSlice
				}
			}

			switch v := value.(type) {
			case bson.M:
				result[key] = ignoreAttributes(v, ignore)
			case []interface{}:
				var slice []interface{}
				for _, item := range v {
					switch item := item.(type) {
					case []map[string]interface{}:
						var slice2 []map[string]interface{}
						for _, item2 := range item {
							slice2 = append(slice2, ignoreAttributes(item2, ignore))
						}
						slice = append(slice, slice2)
					case bson.M:
						slice = append(slice, ignoreAttributes(item, ignore))
					default:
						slice = append(slice, item)
					}
				}
				result[key] = slice
			case []map[string]interface{}:
				var slice []map[string]interface{}
				for _, item := range v {
					slice = append(slice, ignoreAttributes(item, ignore))
				}
				result[key] = slice
			case map[string]interface{}:
				result[key] = ignoreAttributes(v, ignore)
			default:
				result[key] = value
			}
		}
	}

	return result
}

func cleanData(data bson.M, raw bool) bson.M {
	var filteredData bson.M
	if raw {
		filteredData = data
	} else {
		filteredData = ignoreAttributes(data, ignoredAttributes)
	}
	return filteredData
}

func bsonToMap(data bson.M) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		// log.Infof("Key: %s, Value: %v", key, value)
		switch value.(type) {
		case string, int, bool, int64:
			result[key] = value
		case primitive.Binary:
			data := value.(primitive.Binary).Data
			encodedData := base64.StdEncoding.EncodeToString(data)
			result[key] = encodedData
		case bson.A:
			// Handle bson.A (array) by converting to slice of interface{}
			result[key] = convertBsonAToSliceInterface(value.(bson.A))
		case bson.M:
			// Handle nested bson.M by recursively converting to map[string]interface{}
			result[key] = bsonToMap(value.(bson.M))
		case nil:
			result[key] = nil
		default:
			// Handle unknown types (optional: log or return error)
			fmt.Printf("Unknown type encountered while converting key '%s': %T\n", key, value)
		}
	}
	return result
}

func convertBsonAToSliceInterface(data bson.A) []interface{} {
	result := make([]interface{}, 0, len(data))
	for _, element := range data {
		switch element.(type) {
		case int32:
			continue
		case string:
			result = append(result, element.(string))
		default:
			result = append(result, bsonToMap(element.(bson.M)))
		}
	}
	return result
}

func convertMxMicroflowNodeToMap(node *MxMicroflowNode) map[string]interface{} {
	c := make(map[string]interface{})
	c["Type"] = node.Type
	c["ID"] = node.ID
	c["Attributes"] = node.Attributes
	return c
}
