package mpr

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestFastPathHint(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mxunitPath := filepath.Join(tmpDir, "unit.mxunit")
	contents := []byte("mxunit-contents")
	if err := os.WriteFile(mxunitPath, contents, 0644); err != nil {
		t.Fatalf("write mxunit: %v", err)
	}
	info, err := os.Stat(mxunitPath)
	if err != nil {
		t.Fatalf("stat mxunit: %v", err)
	}

	entry := exportManifestEntry{
		ModTimeNs: info.ModTime().UnixNano(),
		FileSize:  info.Size(),
	}
	if !manifestFastPathHint(entry, mxunitPath) {
		t.Fatal("expected matching mtime/size to pass fast path hint")
	}

	entry.FileSize = info.Size() + 1
	if manifestFastPathHint(entry, mxunitPath) {
		t.Fatal("expected size mismatch to fail fast path hint")
	}

	entry.FileSize = info.Size()
	entry.ModTimeNs = info.ModTime().UnixNano() - 1
	if manifestFastPathHint(entry, mxunitPath) {
		t.Fatal("expected mtime mismatch to fail fast path hint")
	}
}

func TestResolveDocumentContentsHashUsesDBWhenHintMatches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mxunitPath := filepath.Join(tmpDir, "unit.mxunit")
	if err := os.WriteFile(mxunitPath, []byte("unchanged"), 0644); err != nil {
		t.Fatalf("write mxunit: %v", err)
	}
	info, err := os.Stat(mxunitPath)
	if err != nil {
		t.Fatalf("stat mxunit: %v", err)
	}

	dbHash := "db-hash-from-sqlite"
	entry := exportManifestEntry{
		ContentsHash: dbHash,
		ModTimeNs:    info.ModTime().UnixNano(),
		FileSize:     info.Size(),
	}

	got, err := resolveDocumentContentsHash(dbHash, entry, mxunitPath, true)
	if err != nil {
		t.Fatalf("resolveDocumentContentsHash() error: %v", err)
	}
	if got != dbHash {
		t.Fatalf("expected db hash %q, got %q", dbHash, got)
	}
}

func TestResolveDocumentContentsHashReadsFileWhenHintMismatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mxunitPath := filepath.Join(tmpDir, "unit.mxunit")
	contents := []byte("changed-mxunit")
	if err := os.WriteFile(mxunitPath, contents, 0644); err != nil {
		t.Fatalf("write mxunit: %v", err)
	}
	sum := sha256.Sum256(contents)
	wantHash := hex.EncodeToString(sum[:])

	entry := exportManifestEntry{
		ContentsHash: "stale-db-hash",
		ModTimeNs:    1,
		FileSize:     1,
	}

	got, err := resolveDocumentContentsHash("stale-db-hash", entry, mxunitPath, true)
	if err != nil {
		t.Fatalf("resolveDocumentContentsHash() error: %v", err)
	}
	if got != wantHash {
		t.Fatalf("expected file hash %q, got %q", wantHash, got)
	}
}

func TestTryFastSkipExportRejectsMtimeMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	mxunitPath := filepath.Join(tmpDir, "unit.mxunit")
	if err := os.WriteFile(mxunitPath, []byte("page"), 0644); err != nil {
		t.Fatalf("write mxunit: %v", err)
	}

	plan := &exportPlan{
		mxunitPaths: map[string]string{"unit-1": mxunitPath},
		manifest: &exportManifest{
			Entries: map[string]exportManifestEntry{
				"unit-1": {
					Name:         "Home_Web",
					ContentsHash: "hash",
					RelativePath: "MyFirstModule/Home_Web.Forms$Page.yaml",
					ModTimeNs:    1,
					FileSize:     1,
				},
			},
		},
	}

	skipped, err := plan.tryFastSkipExport(exportDocumentDescriptor{
		UnitID:       "unit-1",
		Name:         "Home_Web",
		ContentsHash: "hash",
	}, tmpDir, false)
	if err != nil {
		t.Fatalf("tryFastSkipExport() error: %v", err)
	}
	if skipped {
		t.Fatal("expected fast skip to be rejected when mxunit mtime/size changed")
	}
}
