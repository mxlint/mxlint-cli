package mpr

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var log = logrus.New()

func SetLogger(logger *logrus.Logger) {
	log = logger
}

const mendixCacheDirName = ".mendix-cache"

// skipMendixCacheDir returns filepath.SkipDir for .mendix-cache so walks never
// descend into Mendix-managed cache content.
func skipMendixCacheDir(path string, info os.FileInfo) error {
	if info.IsDir() && info.Name() == mendixCacheDirName {
		log.Debugf("Skipping system managed directory %s", path)
		return filepath.SkipDir
	}
	return nil
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

// stripIgnoredAttributes removes ignored keys and normalises primitive.A
// slices in-place, avoiding a full deep copy.
func stripIgnoredAttributes(data map[string]interface{}, ignore []string) {
	for _, key := range ignore {
		delete(data, key)
	}
	for key, value := range data {
		switch v := value.(type) {
		case primitive.A:
			s := []interface{}(v)
			if len(s) > 0 {
				if _, ok := s[0].(int32); ok {
					s = s[1:]
				}
			}
			data[key] = s
			stripSliceItems(s, ignore)
		case []interface{}:
			if len(v) > 0 {
				if _, ok := v[0].(int32); ok {
					v = v[1:]
					data[key] = v
				}
			}
			stripSliceItems(v, ignore)
		case bson.M:
			stripIgnoredAttributes(v, ignore)
		case map[string]interface{}:
			stripIgnoredAttributes(v, ignore)
		case []bson.M:
			for _, item := range v {
				stripIgnoredAttributes(item, ignore)
			}
		case []map[string]interface{}:
			for _, item := range v {
				stripIgnoredAttributes(item, ignore)
			}
		}
	}
}

func stripSliceItems(s []interface{}, ignore []string) {
	for _, item := range s {
		switch m := item.(type) {
		case bson.M:
			stripIgnoredAttributes(m, ignore)
		case map[string]interface{}:
			stripIgnoredAttributes(m, ignore)
		}
	}
}

func cleanData(data bson.M, raw bool) bson.M {
	if !raw {
		stripIgnoredAttributes(data, ignoredAttributes)
	}
	return data
}

func bsonToMap(data bson.M) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		switch v := value.(type) {
		case string, int, bool, int64:
			result[key] = value
		case primitive.Binary:
			encodedData := base64.StdEncoding.EncodeToString(v.Data)
			result[key] = encodedData
		case bson.A:
			result[key] = convertBsonAToSliceInterface(v)
		case bson.M:
			result[key] = bsonToMap(v)
		case nil:
			result[key] = nil
		default:
			fmt.Printf("Unknown type encountered while converting key '%s': %T\n", key, value)
		}
	}
	return result
}

func convertBsonAToSliceInterface(data bson.A) []interface{} {
	result := make([]interface{}, 0, len(data))
	for _, element := range data {
		switch v := element.(type) {
		case int32:
			continue
		case string:
			result = append(result, v)
		default:
			result = append(result, bsonToMap(element.(bson.M)))
		}
	}
	return result
}
