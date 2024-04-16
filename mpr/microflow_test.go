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
	t.Run("microflow-simple", func(t *testing.T) {
		if err := exportUnits("./../resources/app/App.mpr", "./../tmp", true); err != nil {
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
		sequence := mfObj["MainFunction"].([]interface{})
		if len(sequence) != 5 {
			t.Errorf("Unexpected instructions length. Got: %d", len(sequence))
		}
	})
	t.Run("microflow-with-split", func(t *testing.T) {
		if err := exportUnits("./../resources/app/App.mpr", "./../tmp", true); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSplit.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		// parse file
		var mfObj bson.M
		if err := yaml.Unmarshal(mfFile, &mfObj); err != nil {
			t.Errorf("Failed to unmarshal microflow file")
		}
		// check metadata
		if mfObj["Name"] != "MicroflowSplit" {
			t.Errorf("Unexpected name. Got: %s", mfObj["Name"])
		}

		// check sequence
		sequence := mfObj["MainFunction"].([]interface{})
		if len(sequence) != 5 {
			t.Errorf("Unexpected instructions length. Got: %d", len(sequence))
		}
	})
	t.Run("microflow-split-then-merge", func(t *testing.T) {
		if err := exportUnits("./../resources/app/App.mpr", "./../tmp", true); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSplitThenMerge.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		// parse file
		var mfObj bson.M
		if err := yaml.Unmarshal(mfFile, &mfObj); err != nil {
			t.Errorf("Failed to unmarshal microflow file")
		}
		// check metadata
		if mfObj["Name"] != "MicroflowSplitThenMerge" {
			t.Errorf("Unexpected name. Got: %s", mfObj["Name"])
		}

		// check sequence
		sequence := mfObj["MainFunction"].([]interface{})
		if len(sequence) != 5 {
			t.Errorf("Unexpected instructions length. Got: %d", len(sequence))
		}

		split := sequence[4]
		splits := split.(map[interface{}]interface{})["Splits"].([]interface{})
		if len(splits) != 2 {
			t.Errorf("Unexpected instructions length. Got: %d", len(splits))
		}

		split1 := splits[0].([]interface{})
		if len(split1) != 4 {
			t.Errorf("Unexpected instructions length. Got: %d", len(split1))
		}

		split2 := splits[1].([]interface{})
		if len(split2) != 6 {
			t.Errorf("Unexpected instructions length. Got: %d", len(split2))
		}
	})
}
