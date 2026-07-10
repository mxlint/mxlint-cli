package mpr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const exportManifestVersion = 1

type exportManifestEntry struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	FolderPath   string `json:"folderPath"`
	RelativePath string `json:"relativePath"`
	ContentsHash string `json:"contentsHash"`
	ModTimeNs    int64  `json:"modTimeNs,omitempty"`
	FileSize     int64  `json:"fileSize,omitempty"`
}

type exportManifest struct {
	Version int                            `json:"version"`
	Entries map[string]exportManifestEntry `json:"entries"`
}

var exportManifestSettings = struct {
	mu   sync.RWMutex
	path string
}{
	path: "",
}

func SetExportManifestPath(path string) {
	exportManifestSettings.mu.Lock()
	defer exportManifestSettings.mu.Unlock()
	exportManifestSettings.path = strings.TrimSpace(path)
}

func getExportManifestPath() string {
	exportManifestSettings.mu.RLock()
	defer exportManifestSettings.mu.RUnlock()
	return exportManifestSettings.path
}

func newExportManifest() *exportManifest {
	return &exportManifest{
		Version: exportManifestVersion,
		Entries: make(map[string]exportManifestEntry),
	}
}

func loadExportManifest(path string) (*exportManifest, error) {
	if path == "" {
		return newExportManifest(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return newExportManifest(), nil
		}
		return nil, fmt.Errorf("error reading export manifest: %w", err)
	}
	manifest := &exportManifest{}
	if err := json.Unmarshal(data, manifest); err != nil {
		return nil, fmt.Errorf("error parsing export manifest: %w", err)
	}
	if manifest.Entries == nil {
		manifest.Entries = make(map[string]exportManifestEntry)
	}
	return manifest, nil
}

func saveExportManifest(path string, manifest *exportManifest) error {
	if path == "" || manifest == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("error creating export manifest directory: %w", err)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling export manifest: %w", err)
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("error writing export manifest: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("error saving export manifest: %w", err)
	}
	return nil
}

func (m *exportManifest) entryFor(unitID, contentsHash string) (exportManifestEntry, bool) {
	if m == nil {
		return exportManifestEntry{}, false
	}
	entry, ok := m.Entries[unitID]
	if !ok || entry.ContentsHash != contentsHash {
		return exportManifestEntry{}, false
	}
	return entry, true
}

func mxunitFileStat(path string) (modTimeNs, size int64, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}
	return info.ModTime().UnixNano(), info.Size(), nil
}

// manifestFastPathHint returns true when mtime and size match the manifest entry.
func manifestFastPathHint(entry exportManifestEntry, mxunitPath string) bool {
	if entry.ModTimeNs == 0 && entry.FileSize == 0 {
		return true
	}
	modTimeNs, size, err := mxunitFileStat(mxunitPath)
	if err != nil {
		return false
	}
	if entry.FileSize != 0 && entry.FileSize != size {
		return false
	}
	if entry.ModTimeNs != 0 && entry.ModTimeNs != modTimeNs {
		return false
	}
	return true
}

// resolveDocumentContentsHash returns the hash used for export caching.
// When a manifest entry exists but the mxunit mtime/size changed, the SQLite
// ContentsHash may be stale; read the mxunit file for the authoritative hash.
func resolveDocumentContentsHash(dbHash string, entry exportManifestEntry, mxunitPath string, hasManifestEntry bool) (string, error) {
	if !hasManifestEntry || mxunitPath == "" {
		return dbHash, nil
	}
	if manifestFastPathHint(entry, mxunitPath) {
		return dbHash, nil
	}
	fileHash, err := hashMxUnitAtPath(mxunitPath)
	if err != nil {
		return "", err
	}
	return fileHash, nil
}
