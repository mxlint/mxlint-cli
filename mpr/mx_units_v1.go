package mpr

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	_ "github.com/glebarez/go-sqlite"
)

func readMxUnitsV1(inputDirectory string) ([]MxUnit, error) {

	mprPath, err := getMprPath(inputDirectory)
	if err != nil {
		return nil, err
	}
	return getMxUnitsV1(mprPath)

}

func getMprPath(inputDirectory string) (string, error) {
	var mprPath string
	found := false

	err := filepath.Walk(inputDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, ".mendix-cache") {
			log.Debugf("Skipping system managed file %s", path)
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".mpr") {
			mprPath = path
			found = true
			return filepath.SkipDir // Stop walking once we find an MPR file
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if !found {
		return "", fmt.Errorf("no .mpr file found in directory: %s", inputDirectory)
	}

	return mprPath, nil
}

func getMxUnitsV1(MPRFilePath string) ([]MxUnit, error) {
	db, err := sql.Open("sqlite", MPRFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT UnitID, ContainerID, ContainmentName, Contents FROM Unit")
	if err != nil {
		return nil, fmt.Errorf("error querying units: %v", err)
	}
	defer rows.Close()

	units := make([]MxUnit, 0)

	for rows.Next() {
		var containmentName string
		var unitID, containerID, contents []byte
		if err := rows.Scan(&unitID, &containerID, &containmentName, &contents); err != nil {
			return nil, fmt.Errorf("error scanning unit: %v", err)
		}

		var result bson.M

		err := bson.Unmarshal(contents, &result)
		if err != nil {
			return nil, fmt.Errorf("error parsing unit: %v", err)
		}

		// create unit object
		myUnit := MxUnit{
			UnitID:          base64.StdEncoding.EncodeToString(unitID),
			ContainerID:     base64.StdEncoding.EncodeToString(containerID),
			ContainmentName: containmentName,
			Contents:        result,
		}

		units = append(units, myUnit)
	}
	return units, nil
}
