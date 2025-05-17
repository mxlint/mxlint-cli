package mpr

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Varjelus/dirsync"
	"github.com/ghodss/yaml"

	_ "github.com/glebarez/go-sqlite"
)

func ExportModel(inputDirectory string, outputDirectory string, raw bool, mode string, appstore bool) error {

	// create tmp directory in user tmp directory
	tmpDir := filepath.Join(os.TempDir(), "mxlint")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("error creating tmp directory: %v", err)
	}
	log.Debugf("Created tmp directory: %s", tmpDir)
	defer os.RemoveAll(tmpDir)

	log.Infof("Exporting to %s", tmpDir)

	units, err := getMxUnits(inputDirectory)
	if err != nil {
		log.Errorf("Failed to parse MxUnits: %s", err)
		return err
	}

	modules := getMxModules(units)

	if err := exportMetadata(inputDirectory, tmpDir, modules); err != nil {
		return fmt.Errorf("error exporting metadata: %v", err)
	}

	if err := exportUnits(inputDirectory, tmpDir, raw, mode); err != nil {
		return fmt.Errorf("error exporting units: %v", err)
	}

	// remove output directory if it exists
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDirectory, 0755); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}

	// copy tmp directory to output directory
	err = dirsync.Sync(tmpDir, outputDirectory)
	if err != nil {
		return fmt.Errorf("error moving tmp directory to output directory: %v", err)
	}

	if !appstore {
		// remove appstore modules
		removeAppstoreModules(outputDirectory, modules)
	}

	log.Infof("Completed model export")
	return nil
}

func getMprVersion(MPRFilePath string) (int, error) {

	db, err := sql.Open("sqlite", MPRFilePath)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT _FormatVersion FROM _MetaData")
	if err != nil {
		return 1, nil
	}

	defer rows.Close()

	if !rows.Next() {
		return -1, fmt.Errorf("no metadata found")
	}

	var FormatVersion string
	if err := rows.Scan(&FormatVersion); err != nil {
		return -1, fmt.Errorf("error reading _formatVersion from metadata: %v", err)
	}

	if strings.Compare(FormatVersion, "2") == 0 {
		return 2, nil
	} else {
		return 1, nil
	}
}

func exportMetadata(inputDirectory string, outputDirectory string, modules []MxModule) error {

	mprPath, err := getMprPath(inputDirectory)
	if err != nil {
		return err
	}

	mprVersion, err := getMprVersion(mprPath)
	if err != nil {
		return fmt.Errorf("error getting mpr version: %v", err)
	}

	log.Infof("MPR version detected: %d", mprVersion)

	db, err := sql.Open("sqlite", mprPath)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SELECT _ProductVersion, _BuildVersion FROM _MetaData")

	if err != nil {
		return fmt.Errorf("error querying units: %v", err)
	}

	log.Debugf("Exporting metadata")
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
		Modules:        modules,
	}

	// write metadata to file
	metadataYAML, err := yaml.Marshal(metadataObj)
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %v", err)
	}

	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDirectory, 0755); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}
	metadataFileName := filepath.Join(outputDirectory, "Metadata.yaml")

	if err := os.WriteFile(metadataFileName, metadataYAML, 0644); err != nil {
		return fmt.Errorf("error writing metadata file: %v", err)
	}

	return nil

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
			log.Debugf("Unit: %v", unit)
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
				Name:       ".",
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

func getMxDocuments(units []MxUnit, folders []MxFolder, mode string) ([]MxDocument, error) {
	var documents []MxDocument
	documentTypes := []string{"ProjectDocuments", "DomainModel", "ModuleSettings", "ModuleSecurity", "Documents"}

	for _, unit := range units {
		if Contains(documentTypes, unit.ContainmentName) {
			log.Debugf("Unit: %v", unit)
			var name = ""
			if unit.Contents["Name"] != nil {
				name = unit.Contents["Name"].(string)
			}

			myDocument := MxDocument{
				Name:       name,
				Type:       unit.Contents["$Type"].(string),
				Path:       getMxDocumentPath(unit.ContainerID, folders),
				Attributes: unit.Contents,
			}

			if mode == "advanced" && unit.Contents["$Type"] == "Microflows$Microflow" {
				myDocument = transformMicroflow(myDocument)
			}
			documents = append(documents, myDocument)
		}
	}
	log.Infof("Found %d documents", len(documents))
	return documents, nil
}

func exportUnits(inputDirectory string, outputDirectory string, raw bool, mode string) error {

	units, err := getMxUnits(inputDirectory)
	if err != nil {
		return fmt.Errorf("error getting units: %v", err)
	}
	folders, err := getMxFolders(units)
	if err != nil {
		return fmt.Errorf("error getting folders: %v", err)
	}
	documents, err := getMxDocuments(units, folders, mode)
	if err != nil {
		return fmt.Errorf("error getting documents: %v", err)
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
		fname := fmt.Sprintf("%s.%s.yaml", document.Name, document.Type)
		if document.Name == "" {
			fname = fmt.Sprintf("%s.yaml", document.Type)
		}
		attributes := cleanData(document.Attributes, raw)
		err = writeFile(filepath.Join(directory, fname), attributes)
		if err != nil {
			log.Errorf("Error writing file: %v", err)
			return err
		}
	}

	return nil

}

func writeFile(filepath string, contents map[string]interface{}) error {
	log.Debugf("Writing file %s", filepath)
	yamlstring, err := yaml.Marshal(contents)
	if err != nil {
		return fmt.Errorf("error marshaling: %v", err)
	}

	if err := os.WriteFile(filepath, yamlstring, 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}

func getMxUnits(inputDirectory string) ([]MxUnit, error) {
	mprPath, err := getMprPath(inputDirectory)
	if err != nil {
		return nil, err
	}
	mprVersion, err := getMprVersion(mprPath)
	if err != nil {
		return nil, err
	}

	if mprVersion == 2 {
		return readMxUnitsV2(inputDirectory)
	} else {
		return readMxUnitsV1(inputDirectory)
	}
}

// removeAppstoreModules removes appstore modules from the temporary directory
func removeAppstoreModules(tmpDir string, modules []MxModule) error {
	for _, module := range modules {
		// Check if module is an appstore module by looking at its attributes
		if isAppstoreModule(module) {
			moduleDir := filepath.Join(tmpDir, module.Name)
			log.Infof("Discarding appstore module: %s", moduleDir)
			if err := os.RemoveAll(moduleDir); err != nil {
				return fmt.Errorf("error removing appstore module %s: %v", module.Name, err)
			}
		}
	}
	return nil
}

// isAppstoreModule checks if a module is an appstore module based on its attributes
func isAppstoreModule(module MxModule) bool {
	// Check for appstore module indicators
	if module.Attributes == nil {
		return false
	}

	// Check if module has appstore specific attributes
	if _, ok := module.Attributes["FromAppStore"]; ok {
		fromAppStore := module.Attributes["FromAppStore"].(bool)
		if fromAppStore {
			return true
		}
	}

	return false
}
