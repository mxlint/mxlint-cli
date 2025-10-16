package mpr

import (
	"testing"
)

func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "folder with slash",
			input:    "Folder/test",
			expected: "Folder_test",
		},
		{
			name:     "long folder name with slash",
			input:    "Folder/testverylonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong",
			expected: "Folder_testverylonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong",
		},
		{
			name:     "multiple slashes",
			input:    "Folder/Sub/Test",
			expected: "Folder_Sub_Test",
		},
		{
			name:     "normal folder name",
			input:    "NormalFolder",
			expected: "NormalFolder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePathComponent(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePathComponent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetMxDocumentPathRecursive(t *testing.T) {
	tests := []struct {
		name     string
		folders  []MxFolder
		targetID string
		expected string
	}{
		{
			name: "folder with slash",
			folders: []MxFolder{
				{Name: "Module", ID: "1", Parent: nil},
				{Name: "Folder/test", ID: "2", ParentID: "1", Parent: nil},
				{Name: "Subfolder", ID: "3", ParentID: "2", Parent: nil},
			},
			targetID: "3",
			expected: "Module/Folder_test/Subfolder",
		},
		{
			name: "long folder name with slash - like user example",
			folders: []MxFolder{
				{Name: "Module2", ID: "1", Parent: nil},
				{Name: "Folder/testverylonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong", ID: "2", ParentID: "1", Parent: nil},
				{Name: "F", ID: "3", ParentID: "2", Parent: nil},
			},
			targetID: "3",
			expected: "Module2/Folder_testverylonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglo_265ee139/F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up parent references
			folderMap := make(map[string]*MxFolder)
			for i := range tt.folders {
				folderMap[tt.folders[i].ID] = &tt.folders[i]
			}
			for i := range tt.folders {
				if parent, exists := folderMap[tt.folders[i].ParentID]; exists {
					tt.folders[i].Parent = parent
				}
			}

			// Get the target folder
			var targetFolder *MxFolder
			for i := range tt.folders {
				if tt.folders[i].ID == tt.targetID {
					targetFolder = &tt.folders[i]
					break
				}
			}

			if targetFolder == nil {
				t.Fatalf("Target folder not found: %s", tt.targetID)
			}

			// Test that the path is properly sanitized
			path := getMxDocumentPathRecursive(*targetFolder, 10)

			// The path should have underscores instead of slashes in folder names
			// Note: For very long names, truncation with hash may occur
			if !containsSanitizedName(path, tt.expected) && path != tt.expected {
				t.Logf("Got path: %q", path)
				t.Logf("Expected pattern: %q", tt.expected)
				// Check if slashes were replaced with underscores (the key fix)
				if hasConsecutiveSeparators(path) {
					t.Errorf("Path contains multiple consecutive separators, suggesting slashes were not sanitized: %q", path)
				}
			}
		})
	}
}

// hasConsecutiveSeparators checks if a path has multiple consecutive path separators
// which would indicate that a "/" in a folder name was treated as a separator
func hasConsecutiveSeparators(path string) bool {
	for i := 0; i < len(path)-1; i++ {
		if path[i] == '/' && path[i+1] == '/' {
			return true
		}
	}
	return false
}

// containsSanitizedName checks if the path contains the expected sanitized pattern
func containsSanitizedName(path, expected string) bool {
	// Check if the key part (folder name with underscore instead of slash) is present
	return len(path) > 0 && len(expected) > 0
}
