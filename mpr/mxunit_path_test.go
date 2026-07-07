package mpr

import (
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestUnitIDToMxunitGUID(t *testing.T) {
	t.Parallel()

	// UnitID bytes from App.mpr for mxunit 82d3944d-781d-454f-b408-15318fdc23b8
	unitID, err := hex.DecodeString("4d94d3821d784f45b40815318fdc23b8")
	if err != nil {
		t.Fatalf("decode unit id: %v", err)
	}
	got, err := unitIDToMxunitGUID(unitID)
	if err != nil {
		t.Fatalf("unitIDToMxunitGUID() error: %v", err)
	}
	want := "82d3944d-781d-454f-b408-15318fdc23b8"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestMxunitPathForUnitID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	known := "82d3944d-781d-454f-b408-15318fdc23b8"
	unitID, err := hex.DecodeString("4d94d3821d784f45b40815318fdc23b8")
	if err != nil {
		t.Fatalf("decode unit id: %v", err)
	}
	targetDir := filepath.Join(root, "mprcontents", known[0:2], known[2:4])
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	target := filepath.Join(targetDir, known+".mxunit")
	if err := os.WriteFile(target, []byte("test"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := mxunitPathForUnitID(root, unitID)
	if err != nil {
		t.Fatalf("mxunitPathForUnitID() error: %v", err)
	}
	if got != target {
		t.Fatalf("expected %s, got %s", target, got)
	}
}

func TestContentsHashHexFromDB(t *testing.T) {
	t.Parallel()

	sum, err := hex.DecodeString("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	dbHash := base64.StdEncoding.EncodeToString(sum)
	got, err := contentsHashHexFromDB(dbHash)
	if err != nil {
		t.Fatalf("contentsHashHexFromDB() error: %v", err)
	}
	if got != hex.EncodeToString(sum) {
		t.Fatalf("unexpected hash conversion: %s", got)
	}
}

func TestBuildExportPlanV2FromSQLite(t *testing.T) {
	inputDir := filepath.Clean("./../resources/app-mpr-v2")
	mprPath, err := getMprPath(inputDir)
	if err != nil {
		t.Fatalf("getMprPath: %v", err)
	}
	version, err := getMprVersion(mprPath)
	if err != nil || version != 2 {
		t.Fatalf("expected mpr v2, got version %d err %v", version, err)
	}

	plan, err := buildExportPlanV2(inputDir, mprPath)
	if err != nil {
		t.Fatalf("buildExportPlanV2() error: %v", err)
	}
	if len(plan.Documents) == 0 {
		t.Fatal("expected documents in export plan")
	}
	if len(plan.unitCache) == 0 {
		t.Fatal("expected structure units cached during plan build")
	}
	for _, doc := range plan.Documents {
		if doc.ContentsHash == "" {
			t.Fatalf("document %s missing ContentsHash from SQLite", doc.UnitID)
		}
		if _, ok := plan.mxunitPaths[doc.UnitID]; !ok {
			t.Fatalf("document %s missing mxunit path", doc.UnitID)
		}
		if _, ok := plan.unitCache[doc.UnitID]; ok {
			t.Fatalf("document %s should not be cached during plan build", doc.UnitID)
		}
	}
}
