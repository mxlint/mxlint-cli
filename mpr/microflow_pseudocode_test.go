package mpr

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateMicroflowPseudocode_FromSubMicroflowExample(t *testing.T) {
	data, err := os.ReadFile("../modelsource-v2/Module2/SubMicroflowExample.Microflows$Microflow.yaml")
	if err != nil {
		t.Fatalf("Failed to read sample microflow yaml: %v", err)
	}

	var attrs map[string]interface{}
	if err := yaml.Unmarshal(data, &attrs); err != nil {
		t.Fatalf("Failed to parse sample microflow yaml: %v", err)
	}

	pseudocode, err := generateMicroflowPseudocode("Module2.SubMicroflowExample", attrs)
	if err != nil {
		t.Fatalf("Failed to generate pseudocode: %v", err)
	}

	if !strings.Contains(pseudocode, "IF $counter > 0 THEN") {
		t.Fatalf("Expected pseudocode to include split condition")
	}
	if !strings.Contains(pseudocode, "counter = $counter - 1") {
		t.Fatalf("Expected pseudocode to include decrement action")
	}
}

func TestGenerateMicroflowPseudocode_FromLoopMicroflowExample(t *testing.T) {
	data, err := os.ReadFile("../modelsource-v2/Module2/MicroflowLoopExample.Microflows$Microflow.yaml")
	if err != nil {
		t.Fatalf("Failed to read sample microflow yaml: %v", err)
	}

	var attrs map[string]interface{}
	if err := yaml.Unmarshal(data, &attrs); err != nil {
		t.Fatalf("Failed to parse sample microflow yaml: %v", err)
	}

	pseudocode, err := generateMicroflowPseudocode("Module2.MicroflowLoopExample", attrs)
	if err != nil {
		t.Fatalf("Failed to generate pseudocode: %v", err)
	}

	if !strings.Contains(pseudocode, "MICROFLOW: Module2.MicroflowLoopExample") {
		t.Fatalf("Expected pseudocode header with microflow name")
	}
	if !strings.Contains(pseudocode, "FOR EACH IteratorUser IN UserList") {
		t.Fatalf("Expected pseudocode to include structured FOR EACH loop")
	}
	if !strings.Contains(pseudocode, "call Module2.SubMicroflowExample(User = $IteratorUser)") {
		t.Fatalf("Expected pseudocode to include microflow call in loop branch")
	}
}
