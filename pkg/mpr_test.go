// mpr_test.go
package mpr

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

// TestAdd tests the Add function to ensure it returns correct results.
func TestMPRMetadata(t *testing.T) {
	t.Run("single-mpr", func(t *testing.T) {
		if err := exportMetadata("./../resources/full-app.mpr", "./../tmp"); err != nil {
			t.Errorf("Failed to export metadata from MPR file")
		}

		// open metadata file
		metadataFile, err := os.ReadFile("./../tmp/metadata.yaml")
		if err != nil {
			t.Errorf("Failed to read metadata file")
		}
		// read metadata file
		var metadataObj metadata
		if err := yaml.Unmarshal(metadataFile, &metadataObj); err != nil {
			t.Errorf("Failed to unmarshal metadata file")
		}
		// check metadata
		expectedProductVersion := "9.24.4.11007"
		if metadataObj.ProductVersion != expectedProductVersion {
			t.Errorf("ProductVersion is incorrect. Expected: %s, Got: %s", expectedProductVersion, metadataObj.ProductVersion)
		}
	})
}

func TestMPRUnits(t *testing.T) {
	t.Run("single-mpr", func(t *testing.T) {
		if err := exportUnits("./../resources/full-app.mpr", "./../tmp"); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}
	})
}
