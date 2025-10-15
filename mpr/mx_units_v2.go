package mpr

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func readMxUnitsV2(inputDirectory string) ([]MxUnit, error) {
	mprPath, err := getMprPath(inputDirectory)
	if err != nil {
		return nil, fmt.Errorf("error getting MPR path: %v", err)
	}

	units, err := getMxUnitsV2(mprPath)
	if err != nil {
		return nil, fmt.Errorf("error getting MX units: %v", err)
	}

	// Create map for faster lookups
	unitsMap := make(map[string]int)
	for idx, unit := range units {
		unitsMap[unit.UnitID] = idx
	}

	mprContentsDirectory := filepath.Join(inputDirectory, "mprcontents")
	log.Debugf("Walking directory: %s", mprContentsDirectory)
	err = filepath.Walk(mprContentsDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Error walking path %s: %v", path, err)
			return fmt.Errorf("error walking path %s: %v", path, err)
		}

		if strings.Contains(path, ".mendix-cache") {
			log.Debugf("Skipping system managed file %s", path)
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".mxunit") {
			log.Debugf("Parsing mxunit: %s", path)

			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %v", path, err)
			}

			var result bson.M
			if err := bson.Unmarshal(contents, &result); err != nil {
				log.Errorf("Unable to unmarshal BSON content: %v", err)
				return err
			}

			// Use the BSON result directly instead of marshaling/unmarshaling through YAML
			// This avoids unnecessary memory allocation
			data := map[string]interface{}(result)

			// Debug the structure only when MXLINT_TRACE is set
			if os.Getenv("MXLINT_TRACE") == "true" {
				log.Debugf("BSON data structure for %s: %#v", path, data)
			}

			idData, ok := data["$ID"]
			if !ok {
				return fmt.Errorf("missing $ID field in %s", path)
			}

			log.Debugf("ID data: %#v, type: %T", idData, idData)

			// Handle different possible structures
			var unitID string
			switch id := idData.(type) {
			case primitive.Binary:
				// Native BSON binary type (when not going through YAML)
				if len(id.Data) >= 16 {
					unitID = base64.StdEncoding.EncodeToString(id.Data)
					log.Debugf("Generated base64 ID from primitive.Binary: %s", unitID)
				}
			case map[string]interface{}:
				// Try uppercase "Data" first (original format)
				if dataStr, ok := id["Data"].(string); ok {
					unitID = dataStr
				} else if dataVal, ok := id["data"]; ok {
					log.Debugf("Data field found: %#v, type: %T", dataVal, dataVal)

					// Try to handle different types of binary data representation
					switch dataBytes := dataVal.(type) {
					case primitive.Binary:
						// Handle primitive.Binary nested in map
						if len(dataBytes.Data) >= 16 {
							unitID = base64.StdEncoding.EncodeToString(dataBytes.Data)
							log.Debugf("Generated base64 ID from nested primitive.Binary: %s", unitID)
						}
					case []interface{}:
						var bytes []byte
						for _, b := range dataBytes {
							if num, ok := b.(float64); ok {
								bytes = append(bytes, byte(num))
							} else if num, ok := b.(int); ok {
								bytes = append(bytes, byte(num))
							} else {
								log.Debugf("Unknown byte type: %T value: %#v", b, b)
							}
						}
						if len(bytes) >= 16 {
							// Use base64 encoding to match the format in getMxUnitsV2
							unitID = base64.StdEncoding.EncodeToString(bytes)
							log.Debugf("Generated base64 ID: %s", unitID)
						}
					case []byte:
						if len(dataBytes) >= 16 {
							// Use base64 encoding to match the format in getMxUnitsV2
							unitID = base64.StdEncoding.EncodeToString(dataBytes)
							log.Debugf("Generated base64 ID from []byte: %s", unitID)
						}
					default:
						log.Debugf("Unknown data type: %T", dataBytes)
					}
				} else {
					log.Debugf("ID structure: %+v", id)
					return fmt.Errorf("invalid ID Data field in %s: %+v", path, id)
				}
			case string:
				// Direct string (possible with some YAML parsers)
				unitID = id
			default:
				return fmt.Errorf("unexpected ID type in %s: %T", path, idData)
			}

			if unitID == "" {
				return fmt.Errorf("empty unit ID in %s", path)
			}

			idx, exists := unitsMap[unitID]
			if !exists {
				log.Errorf("Unable to process unit file %s: unit ID %s not found", path, unitID)
				return fmt.Errorf("unable to find unit with ID: %s", unitID)
			}

			units[idx].Contents = result
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error processing mprcontents: %v", err)
	}

	return units, nil
}

func getMxUnitsV2(MPRFilePath string) ([]MxUnit, error) {
	db, err := sql.Open("sqlite", MPRFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT UnitID, ContainerID, ContainmentName FROM Unit")
	if err != nil {
		return nil, fmt.Errorf("error querying units: %v", err)
	}
	defer rows.Close()

	var units []MxUnit

	for rows.Next() {
		var containmentName string
		var unitID, containerID []byte
		if err := rows.Scan(&unitID, &containerID, &containmentName); err != nil {
			return nil, fmt.Errorf("error scanning unit: %v", err)
		}

		unit := MxUnit{
			UnitID:          base64.StdEncoding.EncodeToString(unitID),
			ContainerID:     base64.StdEncoding.EncodeToString(containerID),
			ContainmentName: containmentName,
		}
		log.Debugf("unit: %+v", unit)

		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating units: %v", err)
	}

	return units, nil
}

func findMxUnitByID(units []MxUnit, unitID string) (int, error) {
	for idx, unit := range units {
		if unit.UnitID == unitID {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("unable to find unit with ID: %s", unitID)
}
