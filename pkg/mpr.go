package mpr

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/ghodss/yaml"
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// metadata object
type MxMetadata struct {
	ProductVersion string `yaml:"ProductVersion"`
	BuildVersion   string `yaml:"BuildVersion"`
}

func exportMetadata(MPRFilePath string, outputDirectory string) error {
	db, err := sql.Open("sqlite3", MPRFilePath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT _ProductVersion, _BuildVersion FROM _MetaData")
	if err != nil {
		return fmt.Errorf("error querying metadata: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("no metadata found")
	}

	var productVersion, buildVersion string
	if err := rows.Scan(&productVersion, &buildVersion); err != nil {
		return fmt.Errorf("error scanning metadata: %v", err)
	}

	// create metadata object
	metadataObj := MxMetadata{
		ProductVersion: productVersion,
		BuildVersion:   buildVersion,
	}

	// write metadata to file
	metadataYAML, err := yaml.Marshal(metadataObj)
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %v", err)
	}

	directory := filepath.Join(outputDirectory, "modelsource")
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}
	metadataFileName := filepath.Join(directory, "metadata.yaml")

	if err := os.WriteFile(metadataFileName, metadataYAML, 0644); err != nil {
		return fmt.Errorf("error writing metadata file: %v", err)
	}

	return nil

}

type MxUnit struct {
	UnitID          string                 `yaml:"UnitID"`
	ContainerID     string                 `yaml:"ContainerID"`
	ContainmentName string                 `yaml:"ContainmentName"`
	Contents        map[string]interface{} `yaml:"Contents"`
}

type MxDocument struct {
	Name       string                 `yaml:"Name"`
	Type       string                 `yaml:"Type"`
	Path       string                 `yaml:"Path"`
	Attributes map[string]interface{} `yaml:"Attributes"`
}

type MxDomainModel struct {
	Name       string                 `yaml:"Name"`
	Type       string                 `yaml:"Type"`
	Attributes map[string]interface{} `yaml:"Attributes"`
}

type MxModule struct {
	Name       string                 `yaml:"Name"`
	ID         string                 `yaml:"ID"`
	Attributes map[string]interface{} `yaml:"Attributes"`
}

type MxFolder struct {
	Name       string                 `yaml:"Name"`
	ID         string                 `yaml:"ID"`
	ParentID   string                 `yaml:"ParentID"`
	Parent     *MxFolder              `yaml:"Parent"`
	Attributes map[string]interface{} `yaml:"Attributes"`
}

func ignoreAttributes(data bson.M, ignore []string) bson.M {
	result := make(bson.M)

	for key, value := range data {
		ignoreKey := false

		for _, ignoreAttr := range ignore {
			//fmt.Printf("'%v' == '%v'\n", key, ignoreAttr)
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
					case bson.M:
						slice = append(slice, ignoreAttributes(item, ignore))
					default:
						slice = append(slice, item)
					}
				}
				result[key] = slice
			default:
				result[key] = value
			}
		}
	}

	return result
}

func getMxModules(units []MxUnit) []MxModule {
	modules := make([]MxModule, 0)
	for _, unit := range units {
		if unit.ContainmentName == "Modules" {
			myModule := MxModule{
				Name:       unit.Contents["Name"].(string),
				ID:         unit.UnitID,
				Attributes: unit.Contents,
			}
			modules = append(modules, myModule)
		}
	}
	return modules
}

func getMxFolders(units []MxUnit) ([]MxFolder, error) {
	var folders []MxFolder
	for _, unit := range units {
		if unit.ContainmentName == "Folders" || unit.ContainmentName == "Modules" {
			myFolder := MxFolder{
				Name:       unit.Contents["Name"].(string),
				ID:         unit.UnitID,
				ParentID:   unit.ContainerID,
				Attributes: unit.Contents,
				Parent:     nil,
			}
			folders = append(folders, myFolder)
		} else if unit.ContainmentName == "" {
			myFolder := MxFolder{
				Name:       "modelsource",
				ID:         unit.UnitID,
				ParentID:   unit.ContainerID,
				Attributes: unit.Contents,
				Parent:     nil,
			}
			folders = append(folders, myFolder)
		}
	}

	// Temporary map to hold folder references for easy lookup.
	folderMap := make(map[string]*MxFolder)
	for i := range folders {
		folderMap[folders[i].ID] = &folders[i]
	}

	// Set up the parent references.
	for i, folder := range folders {
		if parent, exists := folderMap[folder.ParentID]; exists && folder.ParentID != folder.ID {
			folders[i].Parent = parent
		}
	}

	return folders, nil
}

func getMxDocumentPathRecursive(folder MxFolder, depth int) string {
	if depth == 0 {
		return ""
	}
	if folder.Parent == nil {
		return folder.Name
	} else {
		return filepath.Join(getMxDocumentPathRecursive(*folder.Parent, depth-1), folder.Name)
	}
}

func getMxDocumentPath(containerID string, folders []MxFolder) string {
	for _, folder := range folders {
		if folder.ID == containerID {
			return getMxDocumentPathRecursive(folder, 10)
		}
	}
	return ""
}

func getMxDocuments(units []MxUnit, folders []MxFolder) ([]MxDocument, error) {
	var documents []MxDocument
	for _, unit := range units {
		if unit.ContainmentName == "Documents" {
			// return nil, fmt.Errorf("error querying documents: %v", unit.ContainmentName)
			myDocument := MxDocument{
				Name:       unit.Contents["Name"].(string),
				Type:       unit.Contents["$Type"].(string),
				Path:       getMxDocumentPath(unit.ContainerID, folders),
				Attributes: unit.Contents,
			}
			documents = append(documents, myDocument)
		}
	}
	return documents, nil
}

func getMxDomainModels(units []MxUnit, folders []MxFolder) ([]MxDomainModel, error) {
	var domainModels []MxDomainModel
	for _, unit := range units {
		if unit.ContainmentName == "DomainModel" {
			// return nil, fmt.Errorf("error querying documents: %v", unit.ContainmentName)
			var moduleName = ""
			for _, folder := range folders {
				if folder.ID == unit.ContainerID {
					moduleName = folder.Name
				}
			}
			myDomainModel := MxDomainModel{
				Name:       moduleName,
				Type:       unit.Contents["$Type"].(string),
				Attributes: unit.Contents,
			}
			domainModels = append(domainModels, myDomainModel)
		}
	}
	return domainModels, nil
}

func exportUnits(MPRFilePath string, outputDirectory string) error {
	db, err := sql.Open("sqlite3", MPRFilePath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT UnitID, ContainerID, ContainmentName, Contents FROM Unit")
	if err != nil {
		return fmt.Errorf("error querying units: %v", err)
	}
	defer rows.Close()

	units := make([]MxUnit, 0)

	for rows.Next() {
		var containmentName string
		var unitID, containerID, contents []byte
		if err := rows.Scan(&unitID, &containerID, &containmentName, &contents); err != nil {
			return fmt.Errorf("error scanning unit: %v", err)
		}

		var result bson.M

		err := bson.Unmarshal(contents, &result)
		if err != nil {
			return fmt.Errorf("error parsing unit: %v", err)
		}

		ignoredAttributes := []string{"$ID", "OriginPointer", "DestinationPointer", "Image", "ImageData"}
		filteredData := ignoreAttributes(result, ignoredAttributes)

		// create unit object
		myUnit := MxUnit{
			UnitID:          base64.StdEncoding.EncodeToString(unitID),
			ContainerID:     base64.StdEncoding.EncodeToString(containerID),
			ContainmentName: containmentName,
			Contents:        filteredData,
		}

		units = append(units, myUnit)
		// metadataFileName := filepath.Join(outputDirectory, fmt.Sprintf("%s.yaml", name))
	}

	// modules := getMxModules(units)
	folders, err := getMxFolders(units)
	if err != nil {
		return fmt.Errorf("error getting folders: %v", err)
	}
	documents, err := getMxDocuments(units, folders)
	if err != nil {
		return fmt.Errorf("error getting documents: %v", err)
	}
	domainModels, err := getMxDomainModels(units, folders)
	if err != nil {
		return fmt.Errorf("error getting domain models: %v", err)
	}
	for _, document := range documents {
		// write document
		directory := filepath.Join(outputDirectory, document.Path)
		// ensure directory exists
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err := os.MkdirAll(directory, 0755); err != nil {
				return fmt.Errorf("error creating directory: %v", err)
			}
		}
		writeFile(filepath.Join(directory, fmt.Sprintf("%s.%s.yaml", document.Name, document.Type)), document.Attributes)
	}
	for _, domainModel := range domainModels {
		// write document
		directory := filepath.Join(outputDirectory, "modelsource", domainModel.Name)
		// ensure directory exists
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err := os.MkdirAll(directory, 0755); err != nil {
				return fmt.Errorf("error creating directory: %v", err)
			}
		}
		writeFile(filepath.Join(directory, fmt.Sprintf("%s.yaml", domainModel.Type)), domainModel.Attributes)
	}
	// fmt.Println(documents)

	return nil

}

func writeFile(filepath string, contents map[string]interface{}) error {
	yamlstring, err := yaml.Marshal(contents)
	if err != nil {
		return fmt.Errorf("error marshaling: %v", err)
	}

	if err := os.WriteFile(filepath, yamlstring, 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}

func processMPRFile(MPRFilePath string) int {
	fmt.Printf("Processing %s\n", MPRFilePath)

	db, err := sql.Open("sqlite3", MPRFilePath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM documents")
	if err != nil {
		log.Fatalf("Error querying documents: %v", err)
	}
	defer rows.Close()

	var counter int
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		var jsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(content), &jsonObj); err != nil {
			log.Fatalf("Error unmarshaling JSON: %v", err)
		}

		yamlData, err := yaml.Marshal(jsonObj)
		if err != nil {
			log.Fatalf("Error marshaling YAML: %v", err)
		}

		fileName := fmt.Sprintf("%s_document_%d.yaml", filepath.Base(MPRFilePath), counter)
		if err := os.WriteFile(fileName, yamlData, 0644); err != nil {
			log.Fatalf("Error writing YAML file: %v", err)
		}
		fmt.Printf("Wrote %s\n", fileName)
		counter++
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
		return -1
	}
	return counter
}
