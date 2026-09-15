package mpr

import (
	"os"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v3"
)

func TestMPRMicroflow(t *testing.T) {
	t.Run("microflow-simple", func(t *testing.T) {
		if _, _, err := exportUnits("./../resources/app-mpr-v1/App.mpr", "./../tmp", true, ""); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSimple.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		var mfObj bson.M
		var node yaml.Node
		if err := yaml.Unmarshal(mfFile, &node); err != nil {
			t.Errorf("Failed to unmarshal microflow file: %v", err)
		}
		if err := node.Decode(&mfObj); err != nil {
			t.Errorf("Failed to decode microflow file: %v", err)
		}
		if mfObj["Name"] != "MicroflowSimple" {
			t.Errorf("Unexpected name. Got: %s", mfObj["Name"])
		}

		pseudocode, ok := mfObj["pseudocode"].(string)
		if !ok || pseudocode == "" {
			t.Errorf("Expected pseudocode to be a non-empty string")
		}
		if ok && pseudocode != "" {
			if !containsText(pseudocode, "BEGIN") || !containsText(pseudocode, "END") {
				t.Errorf("Expected pseudocode to contain BEGIN/END structure")
			}
		}
	})
	t.Run("microflow-simple-has-pseudocode", func(t *testing.T) {
		if _, _, err := exportUnits("./../resources/app-mpr-v1/App.mpr", "./../tmp", true, "MicroflowSimple"); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSimple.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}

		var mfObj bson.M
		var node yaml.Node
		if err := yaml.Unmarshal(mfFile, &node); err != nil {
			t.Errorf("Failed to unmarshal microflow file: %v", err)
		}
		if err := node.Decode(&mfObj); err != nil {
			t.Errorf("Failed to decode microflow file: %v", err)
		}

		pseudocode, ok := mfObj["pseudocode"].(string)
		if !ok || pseudocode == "" {
			t.Errorf("Expected pseudocode to be present in export")
		}
	})
	t.Run("microflow-with-split-has-pseudocode", func(t *testing.T) {
		if _, _, err := exportUnits("./../resources/app-mpr-v1/App.mpr", "./../tmp", true, ""); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSplit.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		var mfObj bson.M
		var node yaml.Node
		if err := yaml.Unmarshal(mfFile, &node); err != nil {
			t.Errorf("Failed to unmarshal microflow file: %v", err)
		}
		if err := node.Decode(&mfObj); err != nil {
			t.Errorf("Failed to decode microflow file: %v", err)
		}
		if mfObj["Name"] != "MicroflowSplit" {
			t.Errorf("Unexpected name. Got: %s", mfObj["Name"])
		}

		pseudocode, ok := mfObj["pseudocode"].(string)
		if !ok || pseudocode == "" {
			t.Errorf("Expected pseudocode to be present in export")
		}
	})
	t.Run("microflow-split-then-merge-has-pseudocode", func(t *testing.T) {
		if _, _, err := exportUnits("./../resources/app-mpr-v1/App.mpr", "./../tmp", true, ""); err != nil {
			t.Errorf("Failed to export units from MPR file")
		}

		mfFile, err := os.ReadFile("./../tmp/MyFirstModule/Folder/MicroflowSplitThenMerge.Microflows$Microflow.yaml")
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}
		var mfObj bson.M
		var node yaml.Node
		if err := yaml.Unmarshal(mfFile, &node); err != nil {
			t.Errorf("Failed to unmarshal microflow file: %v", err)
		}
		if err := node.Decode(&mfObj); err != nil {
			t.Errorf("Failed to decode microflow file: %v", err)
		}
		if mfObj["Name"] != "MicroflowSplitThenMerge" {
			t.Errorf("Unexpected name. Got: %s", mfObj["Name"])
		}

		pseudocode, ok := mfObj["pseudocode"].(string)
		if !ok || pseudocode == "" {
			t.Errorf("Expected pseudocode to be present in export")
		}
	})
}

func containsText(s string, needle string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(needle))
}
