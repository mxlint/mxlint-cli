package mpr

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var documentContainmentTypes = map[string]struct{}{
	"ProjectDocuments": {},
	"DomainModel":      {},
	"ModuleSettings":   {},
	"ModuleSecurity":   {},
	"Documents":        {},
}

type exportDocumentDescriptor struct {
	UnitID      string
	Name        string
	Type        string
	ContainerID string
	Path        string
}

type exportPlan struct {
	Modules   []MxModule
	Documents []exportDocumentDescriptor
	Load      func(unitID string) (bson.M, error)
	Close     func() error
}

func buildExportPlan(inputDirectory string) (*exportPlan, error) {
	mprPath, err := getMprPath(inputDirectory)
	if err != nil {
		return nil, err
	}
	mprVersion, err := getMprVersion(mprPath)
	if err != nil {
		return nil, fmt.Errorf("error getting mpr version: %v", err)
	}
	if mprVersion == 2 {
		return buildExportPlanV2(inputDirectory, mprPath)
	}
	return buildExportPlanV1(mprPath)
}

func buildExportPlanV1(mprPath string) (*exportPlan, error) {
	db, err := sql.Open("sqlite", mprPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	rows, err := db.Query("SELECT UnitID, ContainerID, ContainmentName, Contents FROM Unit")
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("error querying units: %v", err)
	}
	defer rows.Close()

	modules := make([]MxModule, 0)
	folders := make([]MxFolder, 0)
	documents := make([]exportDocumentDescriptor, 0)

	for rows.Next() {
		var containmentName string
		var unitID, containerID, contents []byte
		if err := rows.Scan(&unitID, &containerID, &containmentName, &contents); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("error scanning unit: %v", err)
		}

		var result bson.M
		if err := bson.Unmarshal(contents, &result); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("error parsing unit: %v", err)
		}

		unit := MxUnit{
			UnitID:          base64.StdEncoding.EncodeToString(unitID),
			ContainerID:     base64.StdEncoding.EncodeToString(containerID),
			ContainmentName: containmentName,
		}
		appendUnitDescriptor(unit, result, &modules, &folders, &documents)
	}
	if err := rows.Err(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("error iterating units: %v", err)
	}

	connectFolderParents(folders)
	for i := range documents {
		documents[i].Path = getMxDocumentPath(documents[i].ContainerID, folders)
	}

	loadDocument := func(unitID string) (bson.M, error) {
		rawUnitID, err := base64.StdEncoding.DecodeString(unitID)
		if err != nil {
			return nil, fmt.Errorf("failed decoding unit id %s: %w", unitID, err)
		}

		var contents []byte
		if err := db.QueryRow("SELECT Contents FROM Unit WHERE UnitID = ?", rawUnitID).Scan(&contents); err != nil {
			return nil, fmt.Errorf("failed to query contents for unit %s: %w", unitID, err)
		}
		var result bson.M
		if err := bson.Unmarshal(contents, &result); err != nil {
			return nil, fmt.Errorf("failed to parse contents for unit %s: %w", unitID, err)
		}
		return result, nil
	}

	return &exportPlan{
		Modules:   modules,
		Documents: documents,
		Load:      loadDocument,
		Close:     db.Close,
	}, nil
}

func buildExportPlanV2(inputDirectory string, mprPath string) (*exportPlan, error) {
	units, err := getMxUnitsV2(mprPath)
	if err != nil {
		return nil, err
	}

	unitHeaders := make(map[string]MxUnit, len(units))
	for _, unit := range units {
		unitHeaders[unit.UnitID] = unit
	}

	mxUnitPaths := make(map[string]string, len(units))
	modules := make([]MxModule, 0)
	folders := make([]MxFolder, 0)
	documents := make([]exportDocumentDescriptor, 0)

	mprContentsDirectory := filepath.Join(inputDirectory, "mprcontents")
	err = filepath.Walk(mprContentsDirectory, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := skipMendixCacheDir(path, info); err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".mxunit") {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", path, err)
		}
		var result bson.M
		if err := bson.Unmarshal(contents, &result); err != nil {
			return fmt.Errorf("unable to unmarshal BSON content for %s: %w", path, err)
		}

		unitID, err := extractUnitIDFromRawContents(result, path)
		if err != nil {
			return err
		}
		header, exists := unitHeaders[unitID]
		if !exists {
			return fmt.Errorf("unable to find unit with ID: %s", unitID)
		}
		mxUnitPaths[unitID] = path
		appendUnitDescriptor(header, result, &modules, &folders, &documents)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error processing mprcontents: %v", err)
	}

	connectFolderParents(folders)
	for i := range documents {
		documents[i].Path = getMxDocumentPath(documents[i].ContainerID, folders)
	}

	loadDocument := func(unitID string) (bson.M, error) {
		mxunitPath, ok := mxUnitPaths[unitID]
		if !ok {
			return nil, fmt.Errorf("mxunit path not found for unit %s", unitID)
		}

		contents, err := os.ReadFile(mxunitPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read mxunit %s: %w", mxunitPath, err)
		}
		var result bson.M
		if err := bson.Unmarshal(contents, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal mxunit %s: %w", mxunitPath, err)
		}
		return result, nil
	}

	return &exportPlan{
		Modules:   modules,
		Documents: documents,
		Load:      loadDocument,
		Close: func() error {
			return nil
		},
	}, nil
}

func appendUnitDescriptor(unit MxUnit, contents bson.M, modules *[]MxModule, folders *[]MxFolder, documents *[]exportDocumentDescriptor) {
	if unit.ContainmentName == "Modules" {
		name, _ := contents["Name"].(string)
		fromAppStore, _ := contents["FromAppStore"].(bool)
		appStoreVersion, _ := contents["AppStoreVersion"].(string)
		appStoreGuid, _ := contents["AppStoreGuid"].(string)
		appStoreVersionGuid, _ := contents["AppStoreVersionGuid"].(string)
		appStorePackageId, _ := contents["AppStorePackageId"].(string)
		*modules = append(*modules, MxModule{
			Name:                name,
			ID:                  unit.UnitID,
			FromAppStore:        fromAppStore,
			AppStoreVersion:     appStoreVersion,
			AppStoreGuid:        appStoreGuid,
			AppStoreVersionGuid: appStoreVersionGuid,
			AppStorePackageId:   appStorePackageId,
		})
	}

	if unit.ContainmentName == "Folders" || unit.ContainmentName == "Modules" || unit.ContainmentName == "" {
		name := ""
		if unit.ContainmentName != "" {
			name, _ = contents["Name"].(string)
		}
		*folders = append(*folders, MxFolder{
			Name:     name,
			ID:       unit.UnitID,
			ParentID: unit.ContainerID,
		})
	}

	if _, ok := documentContainmentTypes[unit.ContainmentName]; ok {
		name, _ := contents["Name"].(string)
		docType, _ := contents["$Type"].(string)
		*documents = append(*documents, exportDocumentDescriptor{
			UnitID:      unit.UnitID,
			Name:        name,
			Type:        docType,
			ContainerID: unit.ContainerID,
		})
	}
}

func connectFolderParents(folders []MxFolder) {
	folderMap := make(map[string]*MxFolder)
	for i := range folders {
		folderMap[folders[i].ID] = &folders[i]
	}
	for i, folder := range folders {
		if parent, exists := folderMap[folder.ParentID]; exists && folder.ParentID != folder.ID {
			folders[i].Parent = parent
		}
	}
}

func exportDocumentsFromPlan(plan *exportPlan, outputDirectory string, raw bool, filter string) (int, error) {
	var err error
	var filterRegex *regexp.Regexp
	if filter != "" {
		filterRegex, err = regexp.Compile(filter)
		if err != nil {
			return 0, fmt.Errorf("invalid filter regex pattern: %v", err)
		}
		log.Infof("Applying filter: %s", filter)
	}

	exportedCount := 0
	for _, document := range plan.Documents {
		if filterRegex != nil && !filterRegex.MatchString(document.Name) {
			log.Debugf("Skipping document '%s' (does not match filter)", document.Name)
			continue
		}

		attributes, err := plan.Load(document.UnitID)
		if err != nil {
			return 0, fmt.Errorf("error loading document %s: %w", document.Name, err)
		}

		if docType, _ := attributes["$Type"].(string); docType == microflowDocumentType {
			addMicroflowPseudocode(document.Name, attributes)
		}

		if err := writeDocumentToDisk(document, outputDirectory, cleanData(attributes, raw)); err != nil {
			return 0, err
		}
		exportedCount++
	}

	if filterRegex != nil {
		log.Infof("Exported %d documents matching filter (out of %d total)", exportedCount, len(plan.Documents))
	} else {
		log.Infof("Found %d documents", len(plan.Documents))
	}
	return exportedCount, nil
}

func writeDocumentToDisk(document exportDocumentDescriptor, outputDirectory string, attributes map[string]interface{}) error {
	sanitizedPath := sanitizePath(document.Path)
	if sanitizedPath != document.Path {
		log.Warnf("Sanitized path: '%s' -> '%s'", document.Path, sanitizedPath)
	}

	sanitizedName := sanitizePathComponent(document.Name)
	sanitizedType := sanitizePathComponent(document.Type)
	if sanitizedName != document.Name || sanitizedType != document.Type {
		log.Debugf("Sanitized name: '%s' -> '%s', type: '%s' -> '%s'", document.Name, sanitizedName, document.Type, sanitizedType)
	}

	fname := fmt.Sprintf("%s.%s.yaml", sanitizedName, sanitizedType)
	if document.Name == "" {
		fname = fmt.Sprintf("%s.yaml", sanitizedType)
	}

	adjustedPath, adjustedFilename, err := validatePathLength(outputDirectory, sanitizedPath, fname)
	if err != nil {
		return fmt.Errorf("error adjusting path length: %v", err)
	}

	directory := filepath.Join(outputDirectory, adjustedPath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}

	if err := writeFile(filepath.Join(directory, adjustedFilename), attributes); err != nil {
		log.Errorf("Error writing file: %v", err)
		return err
	}
	return nil
}

func extractUnitIDFromRawContents(data map[string]interface{}, path string) (string, error) {
	idData, ok := data["$ID"]
	if !ok {
		return "", fmt.Errorf("missing $ID field in %s", path)
	}

	switch id := idData.(type) {
	case primitive.Binary:
		if len(id.Data) >= 16 {
			return base64.StdEncoding.EncodeToString(id.Data), nil
		}
	case map[string]interface{}:
		if dataStr, ok := id["Data"].(string); ok && dataStr != "" {
			return dataStr, nil
		}
		if dataVal, ok := id["data"]; ok {
			switch dataBytes := dataVal.(type) {
			case primitive.Binary:
				if len(dataBytes.Data) >= 16 {
					return base64.StdEncoding.EncodeToString(dataBytes.Data), nil
				}
			case []interface{}:
				bytes := make([]byte, 0, len(dataBytes))
				for _, b := range dataBytes {
					if num, ok := b.(float64); ok {
						bytes = append(bytes, byte(num))
						continue
					}
					if num, ok := b.(int); ok {
						bytes = append(bytes, byte(num))
					}
				}
				if len(bytes) >= 16 {
					return base64.StdEncoding.EncodeToString(bytes), nil
				}
			case []byte:
				if len(dataBytes) >= 16 {
					return base64.StdEncoding.EncodeToString(dataBytes), nil
				}
			}
		}
	case string:
		if id != "" {
			return id, nil
		}
	}

	return "", fmt.Errorf("unable to extract unit id from %s", path)
}
