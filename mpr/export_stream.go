package mpr

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"

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
	UnitID       string
	Name         string
	Type         string
	ContainerID  string
	Path         string
	ContentsHash string
}

type cachedUnitContent struct {
	Contents     bson.M
	ContentsHash string
}

type exportPlan struct {
	Modules      []MxModule
	Documents    []exportDocumentDescriptor
	unitCache    map[string]cachedUnitContent
	unitCacheMu  sync.Mutex
	mxunitPaths  map[string]string
	manifest     *exportManifest
	manifestPath string
	manifestMu   sync.Mutex
	Close        func() error
}

func (p *exportPlan) loadDocument(unitID string) (bson.M, error) {
	p.unitCacheMu.Lock()
	if cached, ok := p.unitCache[unitID]; ok {
		p.unitCacheMu.Unlock()
		return cached.Contents, nil
	}
	p.unitCacheMu.Unlock()

	mxunitPath, ok := p.mxunitPaths[unitID]
	if !ok {
		return nil, fmt.Errorf("mxunit path not found for unit %s", unitID)
	}

	_, result, hash, err := readMxUnitAtPath(mxunitPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mxunit %s: %w", mxunitPath, err)
	}

	p.unitCacheMu.Lock()
	p.unitCache[unitID] = cachedUnitContent{Contents: result, ContentsHash: hash}
	p.unitCacheMu.Unlock()
	return result, nil
}

func readMxUnitAtPath(path string) ([]byte, bson.M, string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, "", err
	}
	var result bson.M
	if err := bson.Unmarshal(contents, &result); err != nil {
		return nil, nil, "", fmt.Errorf("unable to unmarshal BSON content for %s: %w", path, err)
	}
	sum := sha256.Sum256(contents)
	return contents, result, hex.EncodeToString(sum[:]), nil
}

func buildExportPlan(inputDirectory string, mprPath string) (*exportPlan, error) {
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
	unitCache := make(map[string]cachedUnitContent)

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

		encodedUnitID := encodeUnitID(unitID)
		contentHash := sha256.Sum256(contents)
		unitCache[encodedUnitID] = cachedUnitContent{
			Contents:     result,
			ContentsHash: hex.EncodeToString(contentHash[:]),
		}

		unit := MxUnit{
			UnitID:          encodedUnitID,
			ContainerID:     encodeContainerID(containerID),
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
		if cached, ok := unitCache[documents[i].UnitID]; ok {
			documents[i].ContentsHash = cached.ContentsHash
		}
	}

	_ = db.Close()
	return &exportPlan{
		Modules:   modules,
		Documents: documents,
		unitCache: unitCache,
		Close: func() error {
			return nil
		},
	}, nil
}

func buildExportPlanV2(inputDirectory string, mprPath string) (*exportPlan, error) {
	manifestPath := getExportManifestPath()
	manifest, err := loadExportManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", mprPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT UnitID, ContainerID, ContainmentName, ContentsHash FROM Unit")
	if err != nil {
		return nil, fmt.Errorf("error querying units: %v", err)
	}
	defer rows.Close()

	unitCache := make(map[string]cachedUnitContent)
	mxunitPaths := make(map[string]string, 256)
	modules := make([]MxModule, 0)
	folders := make([]MxFolder, 0)
	documents := make([]exportDocumentDescriptor, 0)

	for rows.Next() {
		var containmentName string
		var dbContentsHash sql.NullString
		var unitID, containerID []byte
		if err := rows.Scan(&unitID, &containerID, &containmentName, &dbContentsHash); err != nil {
			return nil, fmt.Errorf("error scanning unit: %v", err)
		}

		encodedUnitID := encodeUnitID(unitID)
		contentsHashHex := ""
		if dbContentsHash.Valid {
			contentsHashHex, err = contentsHashHexFromDB(dbContentsHash.String)
			if err != nil {
				return nil, fmt.Errorf("error decoding ContentsHash for unit %s: %w", encodedUnitID, err)
			}
		}

		mxunitPath, err := mxunitPathForUnitID(inputDirectory, unitID)
		if err != nil {
			return nil, fmt.Errorf("error resolving mxunit path for unit %s: %w", encodedUnitID, err)
		}
		mxunitPaths[encodedUnitID] = mxunitPath

		unit := MxUnit{
			UnitID:          encodedUnitID,
			ContainerID:     encodeContainerID(containerID),
			ContainmentName: containmentName,
		}

		if isStructureContainment(containmentName) {
			_, result, fileHash, err := readMxUnitAtPath(mxunitPath)
			if err != nil {
				return nil, fmt.Errorf("error reading structure unit %s: %w", mxunitPath, err)
			}
			hash := contentsHashHex
			if hash == "" {
				hash = fileHash
			}
			unitCache[encodedUnitID] = cachedUnitContent{Contents: result, ContentsHash: hash}
			appendUnitDescriptor(unit, result, &modules, &folders, &documents)
			continue
		}

		if _, ok := documentContainmentTypes[containmentName]; ok {
			doc := exportDocumentDescriptor{
				UnitID:       encodedUnitID,
				ContainerID:  unit.ContainerID,
				ContentsHash: contentsHashHex,
			}
			if entry, ok := manifest.entryFor(encodedUnitID, contentsHashHex); ok {
				doc.Name = entry.Name
				doc.Type = entry.Type
				doc.Path = entry.FolderPath
			}
			documents = append(documents, doc)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating units: %v", err)
	}

	connectFolderParents(folders)
	for i := range documents {
		if documents[i].Path == "" {
			documents[i].Path = getMxDocumentPath(documents[i].ContainerID, folders)
		}
	}

	log.Debugf("Built export plan from SQLite: %d modules, %d documents (%d structure units cached)",
		len(modules), len(documents), len(unitCache))

	return &exportPlan{
		Modules:      modules,
		Documents:    documents,
		unitCache:    unitCache,
		mxunitPaths:  mxunitPaths,
		manifest:     manifest,
		manifestPath: manifestPath,
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

type exportDirCache struct {
	mu   sync.Mutex
	dirs map[string]struct{}
}

func newExportDirCache() *exportDirCache {
	return &exportDirCache{dirs: make(map[string]struct{})}
}

func (c *exportDirCache) mkdir(path string) error {
	c.mu.Lock()
	if _, ok := c.dirs[path]; ok {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	c.mu.Lock()
	c.dirs[path] = struct{}{}
	c.mu.Unlock()
	return nil
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

	if len(plan.Documents) == 0 {
		log.Infof("Found 0 documents")
		return 0, nil
	}

	concurrency := effectiveExportConcurrency(len(plan.Documents))
	dirCache := newExportDirCache()
	jobCh := make(chan exportDocumentDescriptor)
	var wg sync.WaitGroup
	var exportedCount atomic.Int64
	var exportErr atomic.Value

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for document := range jobCh {
				if exportErr.Load() != nil {
					return
				}
				exported, err := exportDocument(plan, document, outputDirectory, raw, filterRegex, dirCache)
				if err != nil {
					exportErr.Store(err)
					return
				}
				if exported {
					exportedCount.Add(1)
				}
			}
		}()
	}

	for _, document := range plan.Documents {
		jobCh <- document
	}
	close(jobCh)
	wg.Wait()

	if stored := exportErr.Load(); stored != nil {
		return 0, stored.(error)
	}

	if err := saveExportManifest(plan.manifestPath, plan.manifest); err != nil {
		log.Warnf("Could not save export manifest: %v", err)
	}

	count := int(exportedCount.Load())
	if filterRegex != nil {
		log.Infof("Exported %d documents matching filter (out of %d total)", count, len(plan.Documents))
	} else {
		log.Infof("Found %d documents", count)
	}
	return count, nil
}

func (p *exportPlan) recordManifestEntry(unitID string, entry exportManifestEntry) {
	p.manifestMu.Lock()
	defer p.manifestMu.Unlock()
	if p.manifest == nil {
		p.manifest = newExportManifest()
	}
	p.manifest.Entries[unitID] = entry
}

func (p *exportPlan) tryFastSkipExport(document exportDocumentDescriptor, outputDirectory string, raw bool) (bool, error) {
	entry, ok := p.manifest.entryFor(document.UnitID, document.ContentsHash)
	if !ok || entry.RelativePath == "" {
		return false, nil
	}
	if mxunitPath, exists := p.mxunitPaths[document.UnitID]; exists {
		if !manifestFastPathHint(entry, mxunitPath) {
			log.Debugf("Manifest mtime/size hint mismatch for %s, verifying via hash", document.Name)
		}
	}

	cachedYAML, found, err := readYAMLFromPersistentCache(document.ContentsHash, raw)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	outPath := filepath.Join(outputDirectory, entry.RelativePath)
	return outputFileMatches(outPath, cachedYAML)
}

func (p *exportPlan) tryExportFromYAMLCache(document exportDocumentDescriptor, outputDirectory string, raw bool, dirCache *exportDirCache) (string, bool, error) {
	entry, ok := p.manifest.entryFor(document.UnitID, document.ContentsHash)
	if !ok || entry.RelativePath == "" {
		return "", false, nil
	}

	cachedYAML, found, err := readYAMLFromPersistentCache(document.ContentsHash, raw)
	if err != nil {
		return "", false, err
	}
	if !found {
		return "", false, nil
	}

	outPath := filepath.Join(outputDirectory, entry.RelativePath)
	if err := dirCache.mkdir(filepath.Dir(outPath)); err != nil {
		return "", false, err
	}
	if same, err := outputFileMatches(outPath, cachedYAML); err != nil {
		return "", false, err
	} else if same {
		return entry.RelativePath, true, nil
	}
	if err := os.WriteFile(outPath, cachedYAML, 0644); err != nil {
		return "", false, fmt.Errorf("error writing cached export file: %w", err)
	}
	return entry.RelativePath, true, nil
}

func exportDocument(plan *exportPlan, document exportDocumentDescriptor, outputDirectory string, raw bool, filterRegex *regexp.Regexp, dirCache *exportDirCache) (bool, error) {
	doc := document

	if doc.Name == "" || doc.Type == "" {
		if entry, ok := plan.manifest.entryFor(doc.UnitID, doc.ContentsHash); ok {
			doc.Name = entry.Name
			doc.Type = entry.Type
			if doc.Path == "" {
				doc.Path = entry.FolderPath
			}
		}
	}

	var attributes bson.M
	if doc.Name == "" || doc.Type == "" {
		var err error
		attributes, err = plan.loadDocument(doc.UnitID)
		if err != nil {
			return false, fmt.Errorf("error loading document metadata %s: %w", doc.UnitID, err)
		}
		if doc.Name == "" {
			doc.Name, _ = attributes["Name"].(string)
		}
		if doc.Type == "" {
			doc.Type, _ = attributes["$Type"].(string)
		}
	}

	if filterRegex != nil && !filterRegex.MatchString(doc.Name) {
		log.Debugf("Skipping document '%s' (does not match filter)", doc.Name)
		return false, nil
	}

	if skipped, err := plan.tryFastSkipExport(doc, outputDirectory, raw); err != nil {
		return false, err
	} else if skipped {
		log.Debugf("Skipping unchanged document '%s'", doc.Name)
		return true, nil
	}

	if relPath, ok, err := plan.tryExportFromYAMLCache(doc, outputDirectory, raw, dirCache); err != nil {
		return false, err
	} else if ok {
		plan.recordManifestEntry(doc.UnitID, manifestEntryForDocument(doc, relPath, plan.mxunitPaths[doc.UnitID]))
		return true, nil
	}

	if attributes == nil {
		var err error
		attributes, err = plan.loadDocument(doc.UnitID)
		if err != nil {
			return false, fmt.Errorf("error loading document %s: %w", doc.Name, err)
		}
	}

	if docType, _ := attributes["$Type"].(string); docType == microflowDocumentType {
		addMicroflowPseudocode(doc.Name, attributes)
	}

	relPath, err := writeDocumentToDisk(doc, outputDirectory, cleanData(attributes, raw), raw, dirCache)
	if err != nil {
		return false, err
	}
	plan.recordManifestEntry(doc.UnitID, manifestEntryForDocument(doc, relPath, plan.mxunitPaths[doc.UnitID]))
	return true, nil
}

func manifestEntryForDocument(document exportDocumentDescriptor, relativePath, mxunitPath string) exportManifestEntry {
	entry := exportManifestEntry{
		Name:         document.Name,
		Type:         document.Type,
		FolderPath:   document.Path,
		RelativePath: relativePath,
		ContentsHash: document.ContentsHash,
	}
	if mxunitPath != "" {
		if modTimeNs, size, err := mxunitFileStat(mxunitPath); err == nil {
			entry.ModTimeNs = modTimeNs
			entry.FileSize = size
		}
	}
	return entry
}

func writeDocumentToDisk(document exportDocumentDescriptor, outputDirectory string, attributes map[string]interface{}, raw bool, dirCache *exportDirCache) (string, error) {
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
		return "", fmt.Errorf("error adjusting path length: %v", err)
	}

	directory := filepath.Join(outputDirectory, adjustedPath)
	if err := dirCache.mkdir(directory); err != nil {
		return "", err
	}

	outPath := filepath.Join(directory, adjustedFilename)
	if err := writeFileWithPersistentCache(outPath, attributes, document.ContentsHash, raw); err != nil {
		log.Errorf("Error writing file: %v", err)
		return "", err
	}

	relPath, err := filepath.Rel(outputDirectory, outPath)
	if err != nil {
		return filepath.Join(adjustedPath, adjustedFilename), nil
	}
	return relPath, nil
}

func extractUnitIDFromRawContents(data map[string]interface{}, path string) (string, error) {
	idData, ok := data["$ID"]
	if !ok {
		return "", fmt.Errorf("missing $ID field in %s", path)
	}

	switch id := idData.(type) {
	case primitive.Binary:
		if len(id.Data) >= 16 {
			return encodeUnitID(id.Data), nil
		}
	case map[string]interface{}:
		if dataStr, ok := id["Data"].(string); ok && dataStr != "" {
			return dataStr, nil
		}
		if dataVal, ok := id["data"]; ok {
			switch dataBytes := dataVal.(type) {
			case primitive.Binary:
				if len(dataBytes.Data) >= 16 {
					return encodeUnitID(dataBytes.Data), nil
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
					return encodeUnitID(bytes), nil
				}
			case []byte:
				if len(dataBytes) >= 16 {
					return encodeUnitID(dataBytes), nil
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
