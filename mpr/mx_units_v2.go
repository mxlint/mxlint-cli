package mpr

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	_ "github.com/glebarez/go-sqlite"
	"go.mongodb.org/mongo-driver/bson"
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
	err = filepath.Walk(mprContentsDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
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

			yamlBytes, err := yaml.Marshal(result)
			if err != nil {
				return fmt.Errorf("error marshaling to YAML: %v", err)
			}

			var data map[string]interface{}
			if err := yaml.Unmarshal(yamlBytes, &data); err != nil {
				return fmt.Errorf("error unmarshaling YAML: %v", err)
			}

			unitID, ok := data["$ID"].(map[string]interface{})["Data"].(string)
			if !ok {
				return fmt.Errorf("invalid or missing unit ID in %s", path)
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
