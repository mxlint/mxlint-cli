package mpr

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

// unitIDToMxunitGUID converts a Mendix UnitID blob to the UUID string used in mxunit paths.
// Mendix stores GUID bytes in little-endian field order (.NET Guid layout).
func unitIDToMxunitGUID(unitID []byte) (string, error) {
	if len(unitID) != 16 {
		return "", fmt.Errorf("expected 16-byte unit id, got %d bytes", len(unitID))
	}
	b := unitID
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		uint32(b[3])<<24|uint32(b[2])<<16|uint32(b[1])<<8|uint32(b[0]),
		uint16(b[5])<<8|uint16(b[4]),
		uint16(b[7])<<8|uint16(b[6]),
		uint16(b[8])<<8|uint16(b[9]),
		uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
	), nil
}

func encodeUnitID(unitID []byte) string {
	return base64.StdEncoding.EncodeToString(unitID)
}

func encodeContainerID(containerID []byte) string {
	if len(containerID) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(containerID)
}

func mxunitPathForUnitID(inputDirectory string, unitID []byte) (string, error) {
	guid, err := unitIDToMxunitGUID(unitID)
	if err != nil {
		return "", err
	}
	return filepath.Join(inputDirectory, "mprcontents", guid[0:2], guid[2:4], guid+".mxunit"), nil
}

// contentsHashHexFromDB converts Mendix SQLite ContentsHash (base64-encoded SHA-256) to hex.
func contentsHashHexFromDB(dbHash string) (string, error) {
	dbHash = strings.TrimSpace(dbHash)
	if dbHash == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(dbHash)
	if err != nil {
		return "", fmt.Errorf("invalid ContentsHash: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func isStructureContainment(containmentName string) bool {
	return containmentName == "Modules" || containmentName == "Folders" || containmentName == ""
}
