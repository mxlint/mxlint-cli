// mpr_test.go
package mpr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestAdd tests the Add function to ensure it returns correct results.
func TestMPRMetadata(t *testing.T) {
	t.Run("single-mpr", func(t *testing.T) {
		if err := exportMetadata("./../resources/app-mpr-v1", "./../tmp", nil); err != nil {
			t.Errorf("Failed to export metadata from MPR file")
		}

		// open metadata file
		metadataFile, err := os.ReadFile("./../tmp/Metadata.yaml")
		if err != nil {
			t.Errorf("Failed to read metadata file")
		}
		// read metadata file
		var metadataObj MxMetadata
		var node yaml.Node
		if err := yaml.Unmarshal(metadataFile, &node); err != nil {
			t.Errorf("Failed to unmarshal metadata file: %v", err)
		}
		if err := node.Decode(&metadataObj); err != nil {
			t.Errorf("Failed to decode metadata file: %v", err)
		}
		// check metadata
		expectedProductVersion := "10.18.3.58900"
		if metadataObj.ProductVersion != expectedProductVersion {
			t.Errorf("ProductVersion is incorrect. Expected: %s, Got: %s", expectedProductVersion, metadataObj.ProductVersion)
		}
	})
}

func TestMPRUnits(t *testing.T) {
	t.Run("single-mpr", func(t *testing.T) {
		if _, err := exportUnits("./../resources/app-mpr-v1", "./../tmp", false, "basic", ""); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}
	})
}

func TestIDAttributesExclusion(t *testing.T) {
	t.Run("verify-id-attributes-excluded", func(t *testing.T) {
		// Export units with ID attributes excluded
		if _, err := exportUnits("./../resources/app-mpr-v1", "./../tmp", false, "basic", ""); err != nil {
			t.Errorf("Failed to export units from MPR file: %v", err)
			return
		}

		// Read a sample unit file
		files, err := os.ReadDir("./../tmp/MyFirstModule")
		if err != nil {
			t.Errorf("Failed to read directory: %v", err)
			return
		}

		// Find a file to test
		var filePath string
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
				filePath = filepath.Join("./../tmp/MyFirstModule", file.Name())
				break
			}
		}

		if filePath == "" {
			t.Errorf("No unit files found to test")
			return
		}

		// Read the file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", filePath, err)
			return
		}

		// Check that ID attributes are excluded
		contentStr := string(content)
		if strings.Contains(contentStr, "\"ID\":") || strings.Contains(contentStr, "\"$ID\":") {
			t.Errorf("ID attributes were not excluded from unit document: %s", filePath)
		}

		// Also check for other ignored attributes
		for _, attr := range ignoredAttributes {
			if strings.Contains(contentStr, fmt.Sprintf("\"%s\":", attr)) {
				t.Errorf("Ignored attribute '%s' was not excluded from unit document: %s", attr, filePath)
			}
		}
	})
}

func TestFilterMetadataOnly(t *testing.T) {
	t.Run("filter-metadata-exact-match", func(t *testing.T) {
		// Clean up test directory
		testDir := "./../tmp-filter-metadata"
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		// Export with filter ^Metadata$
		// According to the code, when filter is "^Metadata$", only metadata is exported, no units
		if err := ExportModel("./../resources/app-mpr-v1", testDir, false, "basic", false, "^Metadata$"); err != nil {
			t.Errorf("Failed to export with Metadata filter: %v", err)
			return
		}

		// Check that Metadata.yaml exists
		metadataPath := filepath.Join(testDir, "Metadata.yaml")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Errorf("Metadata.yaml was not created")
			return
		}

		// Check that no other files/directories were created (since filter is ^Metadata$ and units are skipped)
		entries, err := os.ReadDir(testDir)
		if err != nil {
			t.Errorf("Failed to read test directory: %v", err)
			return
		}

		// Should only have Metadata.yaml
		if len(entries) != 1 {
			t.Errorf("Expected only Metadata.yaml, but found %d entries", len(entries))
			return
		}

		if entries[0].Name() != "Metadata.yaml" {
			t.Errorf("Expected Metadata.yaml, but found %s", entries[0].Name())
		}
	})
}

func TestFilterConstantPattern(t *testing.T) {
	t.Run("filter-constant-pattern", func(t *testing.T) {
		// Clean up test directory
		testDir := "./../tmp-filter-constant"
		os.RemoveAll(testDir)
		defer os.RemoveAll(testDir)

		// Export with filter ^Constant.*
		// This pattern won't match any documents in the test data, so we should get only metadata
		if err := ExportModel("./../resources/app-mpr-v1", testDir, false, "basic", false, "^Constant.*"); err != nil {
			t.Errorf("Failed to export with Constant filter: %v", err)
			return
		}

		// Check that Metadata.yaml exists (always exported)
		metadataPath := filepath.Join(testDir, "Metadata.yaml")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Errorf("Metadata.yaml was not created")
			return
		}

		// Check that no module directories were created (since no documents match the filter)
		entries, err := os.ReadDir(testDir)
		if err != nil {
			t.Errorf("Failed to read test directory: %v", err)
			return
		}

		// Should only have Metadata.yaml since no documents match ^Constant.*
		if len(entries) != 1 {
			t.Errorf("Expected only Metadata.yaml when no documents match filter, but found %d entries", len(entries))
			for _, entry := range entries {
				t.Logf("Found entry: %s", entry.Name())
			}
			return
		}

		if entries[0].Name() != "Metadata.yaml" {
			t.Errorf("Expected Metadata.yaml, but found %s", entries[0].Name())
		}
	})
}
