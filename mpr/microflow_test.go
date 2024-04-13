// microflow_test.go
package mpr

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v2"
)

// TestAdd tests the Add function to ensure it returns correct results.

func TestMPRMicroflow(t *testing.T) {
	t.Run("simple-microflow", func(t *testing.T) {
		if err := exportUnits("./../resources/app/App.mpr", "./../tmp", false); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSimple.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		// parse file
		var mfObj bson.M
		if err := yaml.Unmarshal(mfFile, &mfObj); err != nil {
			t.Errorf("Failed to unmarshal microflow file")
		}
		// check metadata
		if mfObj["Name"] != "MicroflowSimple" {
			t.Errorf("Unexpected name. Got: %s", mfObj["Name"])
		}

		// check sequence
		sequence := mfObj["Sequence"].([]interface{})
		if len(sequence) != 3 {
			t.Errorf("Unexpected sequence length. Got: %d", len(sequence))
		}
	})
}
