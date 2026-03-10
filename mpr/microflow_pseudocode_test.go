package mpr

import (
	"strings"
	"testing"
)

func TestGenerateMicroflowPseudocode_FromSubMicroflowExample(t *testing.T) {
	attrs := subMicroflowFixture()

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
	attrs := loopMicroflowFixture()

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

func subMicroflowFixture() map[string]interface{} {
	return map[string]interface{}{
		"MicroflowReturnType": map[string]interface{}{"$Type": "DataTypes$VoidType"},
		"ObjectCollection": map[string]interface{}{
			"Objects": []interface{}{
				map[string]interface{}{"$ID": "start", "$Type": "Microflows$StartEvent"},
				map[string]interface{}{
					"$ID":   "create",
					"$Type": "Microflows$ActionActivity",
					"Action": map[string]interface{}{
						"$Type":         "Microflows$CreateVariableAction",
						"VariableName":  "counter",
						"InitialValue":  "10",
						"ErrorHandling": "Rollback",
					},
				},
				map[string]interface{}{
					"$ID":   "split",
					"$Type": "Microflows$ExclusiveSplit",
					"SplitCondition": map[string]interface{}{
						"Expression": "$counter > 0",
					},
				},
				map[string]interface{}{
					"$ID":   "decrement",
					"$Type": "Microflows$ActionActivity",
					"Action": map[string]interface{}{
						"$Type":              "Microflows$ChangeVariableAction",
						"ChangeVariableName": "counter",
						"Value":              "$counter - 1",
					},
				},
				map[string]interface{}{"$ID": "end", "$Type": "Microflows$EndEvent"},
			},
		},
		"Flows": []interface{}{
			map[string]interface{}{"$ID": "f1", "OriginPointer": "start", "DestinationPointer": "create", "CaseValues": []interface{}{}},
			map[string]interface{}{"$ID": "f2", "OriginPointer": "create", "DestinationPointer": "split", "CaseValues": []interface{}{}},
			map[string]interface{}{"$ID": "f3", "OriginPointer": "split", "DestinationPointer": "decrement", "CaseValues": []interface{}{map[string]interface{}{"Value": "true"}}},
			map[string]interface{}{"$ID": "f4", "OriginPointer": "split", "DestinationPointer": "end", "CaseValues": []interface{}{map[string]interface{}{"Value": "false"}}},
			map[string]interface{}{"$ID": "f5", "OriginPointer": "decrement", "DestinationPointer": "split", "CaseValues": []interface{}{}},
		},
	}
}

func loopMicroflowFixture() map[string]interface{} {
	return map[string]interface{}{
		"MicroflowReturnType": map[string]interface{}{"$Type": "DataTypes$VoidType"},
		"ObjectCollection": map[string]interface{}{
			"Objects": []interface{}{
				map[string]interface{}{"$ID": "start", "$Type": "Microflows$StartEvent"},
				map[string]interface{}{
					"$ID":   "retrieve",
					"$Type": "Microflows$ActionActivity",
					"Action": map[string]interface{}{
						"$Type":              "Microflows$RetrieveAction",
						"ResultVariableName": "UserList",
						"RetrieveSource": map[string]interface{}{
							"Entity": "System.User",
						},
					},
				},
				map[string]interface{}{
					"$ID":   "loop",
					"$Type": "Microflows$LoopedActivity",
					"LoopSource": map[string]interface{}{
						"ListVariableName": "UserList",
						"VariableName":     "IteratorUser",
					},
					"ObjectCollection": map[string]interface{}{
						"Objects": []interface{}{
							map[string]interface{}{
								"$ID":   "loopSplit",
								"$Type": "Microflows$ExclusiveSplit",
								"SplitCondition": map[string]interface{}{
									"Expression": "$IteratorUser/Blocked",
								},
							},
							map[string]interface{}{
								"$ID":   "loopChange",
								"$Type": "Microflows$ActionActivity",
								"Action": map[string]interface{}{
									"$Type":              "Microflows$ChangeAction",
									"ChangeVariableName": "IteratorUser",
									"Commit":             "Yes",
									"Items": []interface{}{
										map[string]interface{}{
											"$Type":     "Microflows$ChangeActionItem",
											"Attribute": "System.User.Blocked",
											"Value":     "false",
										},
									},
								},
							},
							map[string]interface{}{
								"$ID":   "loopCall",
								"$Type": "Microflows$ActionActivity",
								"Action": map[string]interface{}{
									"$Type": "Microflows$MicroflowCallAction",
									"MicroflowCall": map[string]interface{}{
										"Microflow": "Module2.SubMicroflowExample",
										"ParameterMappings": []interface{}{
											map[string]interface{}{
												"Parameter": "Module2.SubMicroflowExample.User",
												"Argument":  "$IteratorUser",
											},
										},
									},
								},
							},
							map[string]interface{}{
								"$ID":   "loopLog",
								"$Type": "Microflows$ActionActivity",
								"Action": map[string]interface{}{
									"$Type": "Microflows$LogMessageAction",
									"MessageTemplate": map[string]interface{}{
										"Text": "User {1} not blocked",
									},
								},
							},
						},
					},
				},
				map[string]interface{}{"$ID": "end", "$Type": "Microflows$EndEvent"},
			},
		},
		"Flows": []interface{}{
			map[string]interface{}{"$ID": "f1", "OriginPointer": "start", "DestinationPointer": "retrieve", "CaseValues": []interface{}{}},
			map[string]interface{}{"$ID": "f2", "OriginPointer": "retrieve", "DestinationPointer": "loop", "CaseValues": []interface{}{}},
			map[string]interface{}{"$ID": "f3", "OriginPointer": "loop", "DestinationPointer": "end", "CaseValues": []interface{}{}},
			map[string]interface{}{"$ID": "f4", "OriginPointer": "loopSplit", "DestinationPointer": "loopChange", "CaseValues": []interface{}{map[string]interface{}{"Value": "true"}}},
			map[string]interface{}{"$ID": "f5", "OriginPointer": "loopSplit", "DestinationPointer": "loopLog", "CaseValues": []interface{}{map[string]interface{}{"Value": "false"}}},
			map[string]interface{}{"$ID": "f6", "OriginPointer": "loopChange", "DestinationPointer": "loopCall", "CaseValues": []interface{}{}},
		},
	}
}
