// microflow_test.go
package mpr

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v3"
)

// TestAdd tests the Add function to ensure it returns correct results.

func TestMPRMicroflow(t *testing.T) {
	t.Run("microflow-simple", func(t *testing.T) {
		if err := exportUnits("./../resources/app-mpr-v1/App.mpr", "./../tmp", true, "advanced", ""); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSimple.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		// parse file
		var mfObj bson.M
		var node yaml.Node
		if err := yaml.Unmarshal(mfFile, &node); err != nil {
			t.Errorf("Failed to unmarshal microflow file: %v", err)
		}
		if err := node.Decode(&mfObj); err != nil {
			t.Errorf("Failed to decode microflow file: %v", err)
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
		if err := exportUnits("./../resources/app-mpr-v1/App.mpr", "./../tmp", true, "advanced", ""); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSplit.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		// parse file
		var mfObj bson.M
		var node yaml.Node
		if err := yaml.Unmarshal(mfFile, &node); err != nil {
			t.Errorf("Failed to unmarshal microflow file: %v", err)
		}
		if err := node.Decode(&mfObj); err != nil {
			t.Errorf("Failed to decode microflow file: %v", err)
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
		if err := exportUnits("./../resources/app-mpr-v1/App.mpr", "./../tmp", true, "advanced", ""); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSplitThenMerge.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		// parse file
		var mfObj bson.M
		var node yaml.Node
		if err := yaml.Unmarshal(mfFile, &node); err != nil {
			t.Errorf("Failed to unmarshal microflow file: %v", err)
		}
		if err := node.Decode(&mfObj); err != nil {
			t.Errorf("Failed to decode microflow file: %v", err)
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
		splitMap, ok := split.(bson.M)
		if !ok {
			t.Errorf("Expected split to be bson.M, got %T", split)
			return
		}
		splits, ok := splitMap["Splits"].([]interface{})
		if !ok {
			t.Errorf("Expected Splits to be []interface{}, got %T", splitMap["Splits"])
			return
		}
		if len(splits) != 2 {
			t.Errorf("Unexpected instructions length. Got: %d", len(splits))
		}

		split1, ok := splits[0].([]interface{})
		if !ok {
			t.Errorf("Expected splits[0] to be []interface{}, got %T", splits[0])
			return
		}
		if len(split1) != 4 {
			t.Errorf("Unexpected instructions length. Got: %d", len(split1))
		}

		split2, ok := splits[1].([]interface{})
		if !ok {
			t.Errorf("Expected splits[1] to be []interface{}, got %T", splits[1])
			return
		}
		if len(split2) != 4 {
			t.Errorf("Unexpected instructions length. Got: %d", len(split2))
		}
	})
}
