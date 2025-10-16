package mpr

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	_ "github.com/glebarez/go-sqlite"
)

const (
	// Windows is most restrictive at 260 chars
	MaxPathLength = 260

	// Reserve space for base directory and separators
	// This leaves room for output directory path
	SafePathBuffer = 60

	// Maximum safe path length for generated content
	MaxSafePath = MaxPathLength - SafePathBuffer // 200 chars

	// Per-component limit (255 is filesystem limit, but we use lower for safety)
	MaxComponentLength = 100
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

	// Check if we can find an MPR file
	mprPath, err := getMprPath(inputDirectory)
	if err != nil {
		return fmt.Errorf("error finding MPR file: %v", err)
	}
	if mprPath == "" {
		return fmt.Errorf("no MPR file found in directory: %s", inputDirectory)
	}

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

	// Ensure both source and destination directories exist before syncing
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %v", err)
	}

	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		return fmt.Errorf("destination directory does not exist: %v", err)
	}

	// copy tmp directory to output directory
	err = syncDirectories(tmpDir, outputDirectory)
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
			log.Debugf("Unit: %v", unit.ContainmentName)
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
				Name:       "",
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
		return sanitizePathComponent(folder.Name)
	} else {
		return filepath.Join(getMxDocumentPathRecursive(*folder.Parent, depth-1), sanitizePathComponent(folder.Name))
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

// sanitizePathComponent sanitizes a single path component (folder or file name) by replacing
// characters that are invalid in file systems with underscores
func sanitizePathComponent(name string) string {
	if name == "" {
		return name
	}

	// Characters that are invalid in Windows: < > : " / \ | ? *
	// Also handle control characters and other problematic characters
	invalidChars := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*", ".."}

	sanitized := name
	for _, char := range invalidChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}

	// Replace control characters (ASCII 0-31) and DEL (127)
	// Also replace newlines, carriage returns, tabs, and null bytes
	result := strings.Builder{}
	for _, r := range sanitized {
		if r < 32 || r == 127 {
			result.WriteRune('_')
		} else {
			result.WriteRune(r)
		}
	}
	sanitized = result.String()

	// Trim leading/trailing spaces and dots (problematic on Windows)
	sanitized = strings.Trim(sanitized, " .")

	// If the name is now empty after trimming, use a default
	if sanitized == "" {
		sanitized = "unnamed"
	}

	// Check for Windows reserved names (CON, PRN, AUX, NUL, COM1-9, LPT1-9)
	// These are case-insensitive on Windows
	upper := strings.ToUpper(sanitized)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL"}
	for _, reserved := range reservedNames {
		if upper == reserved {
			sanitized = "_" + sanitized
			break
		}
	}
	// Check COM1-COM9 and LPT1-LPT9
	if len(upper) == 4 {
		prefix := upper[:3]
		if (prefix == "COM" || prefix == "LPT") && upper[3] >= '1' && upper[3] <= '9' {
			sanitized = "_" + sanitized
		}
	}

	// Enforce maximum component length
	if len(sanitized) > MaxComponentLength {
		sanitized = truncatePathComponent(sanitized, MaxComponentLength)
	}

	return sanitized
}

// sanitizePath sanitizes a full path by sanitizing each component
func sanitizePath(path string) string {
	// Split the path into components
	components := strings.Split(path, string(filepath.Separator))

	// Sanitize each component
	for i, component := range components {
		components[i] = sanitizePathComponent(component)
	}

	// Rejoin the path
	return filepath.Join(components...)
}

// truncatePathComponent truncates a path component to maxLen while maintaining uniqueness
// If truncation is needed, it appends a hash to ensure uniqueness
func truncatePathComponent(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}

	// Create a short hash of the full name for uniqueness
	hash := sha256.Sum256([]byte(name))
	hashStr := hex.EncodeToString(hash[:])[:8] // Use first 8 chars of hash

	// Reserve space for hash and separator
	maxTextLen := maxLen - len(hashStr) - 1 // -1 for underscore

	if maxTextLen < 1 {
		// If maxLen is too small, just use the hash
		return hashStr[:maxLen]
	}

	// Truncate and append hash
	return name[:maxTextLen] + "_" + hashStr
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// validatePathLength checks if the full path would exceed limits and adjusts if needed
func validatePathLength(basePath string, relativePath string, filename string) (string, string, error) {
	fullPath := filepath.Join(basePath, relativePath, filename)

	if len(fullPath) <= MaxSafePath {
		return relativePath, filename, nil
	}

	// Path is too long, need to adjust
	log.Warnf("Path exceeds safe length (%d chars): %s", len(fullPath), fullPath)

	// Strategy: Truncate path components starting from the deepest
	components := strings.Split(relativePath, string(filepath.Separator))

	// Calculate how much we need to save
	excess := len(fullPath) - MaxSafePath

	// Try to shorten components from the end (deepest folders)
	for i := len(components) - 1; i >= 0 && excess > 0; i-- {
		oldLen := len(components[i])
		targetLen := max(10, oldLen-excess) // Keep at least 10 chars if possible

		if oldLen > targetLen {
			components[i] = truncatePathComponent(components[i], targetLen)
			excess -= (oldLen - len(components[i]))
		}
	}

	// If still too long, truncate the filename
	if excess > 0 {
		maxFilenameLen := len(filename) - excess - 10 // Reserve 10 chars minimum
		if maxFilenameLen < 10 {
			maxFilenameLen = 10
		}
		filename = truncatePathComponent(filename, maxFilenameLen)
	}

	newRelativePath := filepath.Join(components...)
	newFullPath := filepath.Join(basePath, newRelativePath, filename)

	log.Warnf("Adjusted path from %d to %d chars", len(fullPath), len(newFullPath))

	return newRelativePath, filename, nil
}

func getMxDocuments(units []MxUnit, folders []MxFolder, mode string) ([]MxDocument, error) {
	var documents []MxDocument
	documentTypes := []string{"ProjectDocuments", "DomainModel", "ModuleSettings", "ModuleSecurity", "Documents"}

	for _, unit := range units {
		if Contains(documentTypes, unit.ContainmentName) {
			log.Debugf("Unit: %v", unit.ContainmentName)
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
	log.Debugf("Exporting units from %s to %s", inputDirectory, outputDirectory)

	units, err := getMxUnits(inputDirectory)
	if err != nil {
		log.Errorf("Error getting units: %v", err)
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
		// Sanitize the document path to handle invalid characters
		sanitizedPath := sanitizePath(document.Path)
		if sanitizedPath != document.Path {
			log.Warnf("Sanitized path: '%s' -> '%s'", document.Path, sanitizedPath)
		}

		// Sanitize the document name to handle invalid characters
		sanitizedName := sanitizePathComponent(document.Name)
		sanitizedType := sanitizePathComponent(document.Type)
		if sanitizedName != document.Name || sanitizedType != document.Type {
			log.Debugf("Sanitized name: '%s' -> '%s', type: '%s' -> '%s'", document.Name, sanitizedName, document.Type, sanitizedType)
		}

		fname := fmt.Sprintf("%s.%s.yaml", sanitizedName, sanitizedType)
		if document.Name == "" {
			fname = fmt.Sprintf("%s.yaml", sanitizedType)
		}

		// Validate and adjust path length to prevent exceeding OS limits
		adjustedPath, adjustedFilename, err := validatePathLength(outputDirectory, sanitizedPath, fname)
		if err != nil {
			return fmt.Errorf("error adjusting path length: %v", err)
		}

		directory := filepath.Join(outputDirectory, adjustedPath)

		// ensure directory exists
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err := os.MkdirAll(directory, 0755); err != nil {
				return fmt.Errorf("error creating directory: %v", err)
			}
		}

		attributes := cleanData(document.Attributes, raw)
		err = writeFile(filepath.Join(directory, adjustedFilename), attributes)
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
		log.Errorf("Error getting MPR path: %v", err)
		return nil, err
	}
	mprVersion, err := getMprVersion(mprPath)
	if err != nil {
		log.Errorf("Error getting MPR version: %v", err)
		return nil, err
	}

	log.Debugf("MPR version: %d", mprVersion)
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
			log.Warnf("Ignoring appstore module: %s", moduleDir)
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

// syncDirectories synchronizes the contents of src to dst
func syncDirectories(src, dst string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Walk through the source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from the source root
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Skip the root directory
		if relPath == "." {
			return nil
		}

		// Calculate the destination path
		dstPath := filepath.Join(dst, relPath)

		// If it's a directory, create it in the destination
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy the file
		return copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string, mode os.FileMode) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	// Copy the contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	return nil
}
