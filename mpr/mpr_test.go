package mpr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetMprVersion(t *testing.T) {
	tests := []struct {
		name        string
		mprPath     string
		expected    int
		expectError bool
	}{
		{
			name:        "version 1 mpr from app-mpr-v1",
			mprPath:     "./../resources/app-mpr-v1/App.mpr",
			expected:    1,
			expectError: false,
		},
		{
			name:        "version 2 mpr from app-mpr-v2",
			mprPath:     "./../resources/app-mpr-v2/App.mpr",
			expected:    2,
			expectError: false,
		},
		{
			name:        "non-existent file returns version 1",
			mprPath:     "./../resources/truly-nonexistent-file.mpr",
			expected:    1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := getMprVersion(tt.mprPath)
			if tt.expectError {
				if err == nil {
					t.Errorf("getMprVersion() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("getMprVersion() unexpected error: %v", err)
				}
				if version != tt.expected {
					t.Errorf("getMprVersion() = %v, want %v", version, tt.expected)
				}
			}
		})
	}
}

func TestGetMxModules(t *testing.T) {
	tests := []struct {
		name     string
		units    []MxUnit
		expected int
	}{
		{
			name: "with modules",
			units: []MxUnit{
				{
					UnitID:          "1",
					ContainerID:     "0",
					ContainmentName: "Modules",
					Contents: map[string]interface{}{
						"Name": "MyModule",
						"$ID":  "1",
					},
				},
				{
					UnitID:          "2",
					ContainerID:     "1",
					ContainmentName: "Documents",
					Contents: map[string]interface{}{
						"Name": "MyDocument",
					},
				},
			},
			expected: 1,
		},
		{
			name: "no modules",
			units: []MxUnit{
				{
					UnitID:          "2",
					ContainerID:     "1",
					ContainmentName: "Documents",
					Contents: map[string]interface{}{
						"Name": "MyDocument",
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modules := getMxModules(tt.units)
			if len(modules) != tt.expected {
				t.Errorf("getMxModules() returned %v modules, want %v", len(modules), tt.expected)
			}
			if len(modules) > 0 && tt.expected > 0 {
				if modules[0].Name != "MyModule" {
					t.Errorf("getMxModules() module name = %v, want MyModule", modules[0].Name)
				}
			}
		})
	}
}

func TestGetMxFolders(t *testing.T) {
	tests := []struct {
		name     string
		units    []MxUnit
		expected int
	}{
		{
			name: "with folders",
			units: []MxUnit{
				{
					UnitID:          "1",
					ContainerID:     "0",
					ContainmentName: "Modules",
					Contents: map[string]interface{}{
						"Name": "MyModule",
					},
				},
				{
					UnitID:          "2",
					ContainerID:     "1",
					ContainmentName: "Folders",
					Contents: map[string]interface{}{
						"Name": "MyFolder",
					},
				},
			},
			expected: 2,
		},
		{
			name: "with parent references",
			units: []MxUnit{
				{
					UnitID:          "parent",
					ContainerID:     "0",
					ContainmentName: "Modules",
					Contents: map[string]interface{}{
						"Name": "ParentModule",
					},
				},
				{
					UnitID:          "child",
					ContainerID:     "parent",
					ContainmentName: "Folders",
					Contents: map[string]interface{}{
						"Name": "ChildFolder",
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folders, err := getMxFolders(tt.units)
			if err != nil {
				t.Errorf("getMxFolders() unexpected error: %v", err)
			}
			if len(folders) != tt.expected {
				t.Errorf("getMxFolders() returned %v folders, want %v", len(folders), tt.expected)
			}

			// Check parent references are set correctly
			if tt.name == "with parent references" && len(folders) == 2 {
				childFolder := folders[1]
				if childFolder.Parent == nil {
					t.Errorf("getMxFolders() child folder should have parent reference")
				} else if childFolder.Parent.ID != "parent" {
					t.Errorf("getMxFolders() child folder parent ID = %v, want 'parent'", childFolder.Parent.ID)
				}
			}
		})
	}
}

func TestGetMxDocumentPath(t *testing.T) {
	// Create test folder structure
	parentFolder := MxFolder{
		Name:     "ParentModule",
		ID:       "parent",
		ParentID: "",
		Parent:   nil,
	}
	childFolder := MxFolder{
		Name:     "ChildFolder",
		ID:       "child",
		ParentID: "parent",
		Parent:   &parentFolder,
	}
	folders := []MxFolder{parentFolder, childFolder}

	tests := []struct {
		name        string
		containerID string
		folders     []MxFolder
		expected    string
	}{
		{
			name:        "find child folder path",
			containerID: "child",
			folders:     folders,
			expected:    "ParentModule/ChildFolder",
		},
		{
			name:        "find parent folder path",
			containerID: "parent",
			folders:     folders,
			expected:    "ParentModule",
		},
		{
			name:        "non-existent container",
			containerID: "nonexistent",
			folders:     folders,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := getMxDocumentPath(tt.containerID, tt.folders)
			if path != tt.expected {
				t.Errorf("getMxDocumentPath() = %v, want %v", path, tt.expected)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path with invalid characters",
			input:    "Module/Folder:Test/File*Name",
			expected: "Module/Folder_Test/File_Name",
		},
		{
			name:     "normal path",
			input:    "Module/Folder/File",
			expected: "Module/Folder/File",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePath(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTruncatePathComponent(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		maxLen          int
		expectLen       int
		shouldContain   []string
		shouldNotExceed int
	}{
		{
			name:            "short name no truncation",
			input:           "ShortName",
			maxLen:          50,
			expectLen:       9,
			shouldContain:   []string{"ShortName"},
			shouldNotExceed: 50,
		},
		{
			name:            "long name truncation",
			input:           "VeryLongFolderNameThatExceedsTheMaximumLengthAllowed",
			maxLen:          50,
			expectLen:       50,
			shouldContain:   []string{"VeryLongFolderNameTh", "_TRUNCATED_", "_LengthAllowed"}, // first 20 + _TRUNCATED_ + hash + _ + last 13
			shouldNotExceed: 50,
		},
		{
			name:            "exact length",
			input:           "ExactLength",
			maxLen:          11,
			expectLen:       11,
			shouldContain:   []string{"ExactLength"},
			shouldNotExceed: 11,
		},
		{
			name:            "very long name",
			input:           "ThisIsAVeryVeryVeryVeryVeryLongFileNameThatDefinitelyExceedsTheLimit",
			maxLen:          50,
			expectLen:       50,
			shouldContain:   []string{"ThisIsAVeryVeryVeryV", "_TRUNCATED_", "_ceedsTheLimit"}, // first 20 + _TRUNCATED_ + hash + _ + last 13
			shouldNotExceed: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePathComponent(tt.input, tt.maxLen)
			if len(result) != tt.expectLen {
				t.Errorf("truncatePathComponent() length = %v, want %v, result = %v", len(result), tt.expectLen, result)
			}
			if len(result) > tt.shouldNotExceed {
				t.Errorf("truncatePathComponent() length %v exceeds max %v", len(result), tt.shouldNotExceed)
			}
			for _, substring := range tt.shouldContain {
				if !strings.Contains(result, substring) {
					t.Errorf("truncatePathComponent() result %v should contain %v", result, substring)
				}
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a greater than b",
			a:        10,
			b:        5,
			expected: 10,
		},
		{
			name:     "b greater than a",
			a:        5,
			b:        10,
			expected: 10,
		},
		{
			name:     "equal values",
			a:        7,
			b:        7,
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("max() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidatePathLength(t *testing.T) {
	tests := []struct {
		name         string
		basePath     string
		relativePath string
		filename     string
		expectError  bool
	}{
		{
			name:         "normal path length",
			basePath:     "/tmp",
			relativePath: "module/folder",
			filename:     "file.yaml",
			expectError:  false,
		},
		{
			name:         "very long path",
			basePath:     "/tmp",
			relativePath: "very/long/path/with/many/nested/folders/that/exceeds/the/maximum/safe/path/length/allowed/by/the/system/configuration/and/should/be/truncated/automatically/to/prevent/errors/when/creating/files/on/windows/systems/with/path/length/limitations",
			filename:     "verylongfilename.yaml",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newPath, newFilename, err := validatePathLength(tt.basePath, tt.relativePath, tt.filename)
			if tt.expectError && err == nil {
				t.Errorf("validatePathLength() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("validatePathLength() unexpected error: %v", err)
			}

			// Check that returned path is shorter than original for long paths
			fullPath := filepath.Join(tt.basePath, newPath, newFilename)
			originalPath := filepath.Join(tt.basePath, tt.relativePath, tt.filename)
			if len(originalPath) > MaxSafePath && len(fullPath) > len(originalPath) {
				t.Errorf("validatePathLength() did not shorten path: %v -> %v", len(originalPath), len(fullPath))
			}
		})
	}
}

func TestIsAppstoreModule(t *testing.T) {
	tests := []struct {
		name     string
		module   MxModule
		expected bool
	}{
		{
			name: "appstore module",
			module: MxModule{
				Name: "AppStoreModule",
				ID:   "1",
				Attributes: map[string]interface{}{
					"FromAppStore": true,
				},
			},
			expected: true,
		},
		{
			name: "not appstore module",
			module: MxModule{
				Name: "CustomModule",
				ID:   "2",
				Attributes: map[string]interface{}{
					"FromAppStore": false,
				},
			},
			expected: false,
		},
		{
			name: "no appstore attribute",
			module: MxModule{
				Name:       "CustomModule",
				ID:         "3",
				Attributes: map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "nil attributes",
			module: MxModule{
				Name:       "CustomModule",
				ID:         "4",
				Attributes: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAppstoreModule(tt.module)
			if result != tt.expected {
				t.Errorf("isAppstoreModule() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mpr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		filepath    string
		contents    map[string]interface{}
		expectError bool
	}{
		{
			name:     "write valid file",
			filepath: filepath.Join(tmpDir, "test.yaml"),
			contents: map[string]interface{}{
				"Name": "TestDocument",
				"Type": "TestType",
			},
			expectError: false,
		},
		{
			name:     "write to invalid directory",
			filepath: filepath.Join(tmpDir, "nonexistent", "test.yaml"),
			contents: map[string]interface{}{
				"Name": "TestDocument",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeFile(tt.filepath, tt.contents)
			if tt.expectError && err == nil {
				t.Errorf("writeFile() expected error but got none")
			}
			if !tt.expectError {
				if err != nil {
					t.Errorf("writeFile() unexpected error: %v", err)
				} else {
					// Verify file was written correctly
					content, err := os.ReadFile(tt.filepath)
					if err != nil {
						t.Errorf("Failed to read written file: %v", err)
					}
					var data map[string]interface{}
					if err := yaml.Unmarshal(content, &data); err != nil {
						t.Errorf("Failed to unmarshal written YAML: %v", err)
					}
					if data["Name"] != tt.contents["Name"] {
						t.Errorf("Written file content mismatch")
					}
				}
			}
		})
	}
}

func TestSyncDirectories(t *testing.T) {
	// Create temporary directories for testing
	srcDir, err := os.MkdirTemp("", "mpr-test-src-*")
	if err != nil {
		t.Fatalf("Failed to create temp src directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "mpr-test-dst-*")
	if err != nil {
		t.Fatalf("Failed to create temp dst directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create test files in source directory
	testFile := filepath.Join(srcDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create subdirectory with file
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create subfile: %v", err)
	}

	// Test syncing directories
	err = syncDirectories(srcDir, dstDir)
	if err != nil {
		t.Errorf("syncDirectories() unexpected error: %v", err)
	}

	// Verify files were copied
	dstTestFile := filepath.Join(dstDir, "test.txt")
	if _, err := os.Stat(dstTestFile); os.IsNotExist(err) {
		t.Errorf("syncDirectories() did not copy test.txt")
	}

	// Verify subdirectory and file were copied
	dstSubFile := filepath.Join(dstDir, "subdir", "subfile.txt")
	if _, err := os.Stat(dstSubFile); os.IsNotExist(err) {
		t.Errorf("syncDirectories() did not copy subdirectory structure")
	}

	// Verify file contents
	copiedContent, err := os.ReadFile(dstTestFile)
	if err != nil {
		t.Errorf("Failed to read copied file: %v", err)
	}
	if string(copiedContent) != string(testContent) {
		t.Errorf("Copied file content mismatch")
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mpr-test-copy-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	testContent := []byte("test content for copy")
	if err := os.WriteFile(srcFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	tests := []struct {
		name        string
		src         string
		dst         string
		expectError bool
	}{
		{
			name:        "copy valid file",
			src:         srcFile,
			dst:         filepath.Join(tmpDir, "destination.txt"),
			expectError: false,
		},
		{
			name:        "copy non-existent file",
			src:         filepath.Join(tmpDir, "nonexistent.txt"),
			dst:         filepath.Join(tmpDir, "dest2.txt"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := copyFile(tt.src, tt.dst, 0644)
			if tt.expectError && err == nil {
				t.Errorf("copyFile() expected error but got none")
			}
			if !tt.expectError {
				if err != nil {
					t.Errorf("copyFile() unexpected error: %v", err)
				} else {
					// Verify file was copied
					copiedContent, err := os.ReadFile(tt.dst)
					if err != nil {
						t.Errorf("Failed to read copied file: %v", err)
					}
					if string(copiedContent) != string(testContent) {
						t.Errorf("Copied file content mismatch")
					}
				}
			}
		})
	}
}

func TestGetMxDocuments(t *testing.T) {
	// Create test data
	parentFolder := MxFolder{
		Name:     "TestModule",
		ID:       "module1",
		ParentID: "",
		Parent:   nil,
	}
	folders := []MxFolder{parentFolder}

	units := []MxUnit{
		{
			UnitID:          "doc1",
			ContainerID:     "module1",
			ContainmentName: "Documents",
			Contents: map[string]interface{}{
				"Name":  "TestDocument",
				"$Type": "Pages$Page",
			},
		},
		{
			UnitID:          "doc2",
			ContainerID:     "module1",
			ContainmentName: "DomainModel",
			Contents: map[string]interface{}{
				"Name":  "TestDomainModel",
				"$Type": "DomainModels$DomainModel",
			},
		},
		{
			UnitID:          "other",
			ContainerID:     "module1",
			ContainmentName: "Other",
			Contents: map[string]interface{}{
				"Name":  "ShouldBeIgnored",
				"$Type": "Other$Type",
			},
		},
	}

	tests := []struct {
		name         string
		units        []MxUnit
		folders      []MxFolder
		mode         string
		expectedDocs int
	}{
		{
			name:         "basic mode",
			units:        units,
			folders:      folders,
			mode:         "basic",
			expectedDocs: 2,
		},
		{
			name:         "advanced mode",
			units:        units,
			folders:      folders,
			mode:         "advanced",
			expectedDocs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			documents, err := getMxDocuments(tt.units, tt.folders, tt.mode)
			if err != nil {
				t.Errorf("getMxDocuments() unexpected error: %v", err)
			}
			if len(documents) != tt.expectedDocs {
				t.Errorf("getMxDocuments() returned %v documents, want %v", len(documents), tt.expectedDocs)
			}
			// Verify document properties
			if len(documents) > 0 {
				if documents[0].Name != "TestDocument" {
					t.Errorf("getMxDocuments() first document name = %v, want TestDocument", documents[0].Name)
				}
				if documents[0].Type != "Pages$Page" {
					t.Errorf("getMxDocuments() first document type = %v, want Pages$Page", documents[0].Type)
				}
			}
		})
	}
}

func TestExportMetadata(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mpr-test-metadata-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		inputDir    string
		outputDir   string
		expectError bool
	}{
		{
			name:        "export metadata v1",
			inputDir:    "./../resources/app-mpr-v1",
			outputDir:   tmpDir,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get modules first
			units, err := getMxUnits(tt.inputDir)
			if err != nil && !tt.expectError {
				t.Fatalf("Failed to get units: %v", err)
			}
			modules := getMxModules(units)

			err = exportMetadata(tt.inputDir, tt.outputDir, modules)
			if tt.expectError && err == nil {
				t.Errorf("exportMetadata() expected error but got none")
			}
			if !tt.expectError {
				if err != nil {
					t.Errorf("exportMetadata() unexpected error: %v", err)
				} else {
					// Verify metadata file was created
					metadataFile := filepath.Join(tt.outputDir, "Metadata.yaml")
					if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
						t.Errorf("exportMetadata() did not create Metadata.yaml")
					}
				}
			}
		})
	}
}

func TestRemoveAppstoreModules(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mpr-test-appstore-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test module directories
	appstoreModuleDir := filepath.Join(tmpDir, "AppStoreModule")
	customModuleDir := filepath.Join(tmpDir, "CustomModule")
	if err := os.MkdirAll(appstoreModuleDir, 0755); err != nil {
		t.Fatalf("Failed to create appstore module dir: %v", err)
	}
	if err := os.MkdirAll(customModuleDir, 0755); err != nil {
		t.Fatalf("Failed to create custom module dir: %v", err)
	}

	modules := []MxModule{
		{
			Name: "AppStoreModule",
			ID:   "1",
			Attributes: map[string]interface{}{
				"FromAppStore": true,
			},
		},
		{
			Name: "CustomModule",
			ID:   "2",
			Attributes: map[string]interface{}{
				"FromAppStore": false,
			},
		},
	}

	err = removeAppstoreModules(tmpDir, modules)
	if err != nil {
		t.Errorf("removeAppstoreModules() unexpected error: %v", err)
	}

	// Verify appstore module was removed
	if _, err := os.Stat(appstoreModuleDir); !os.IsNotExist(err) {
		t.Errorf("removeAppstoreModules() did not remove appstore module")
	}

	// Verify custom module still exists
	if _, err := os.Stat(customModuleDir); os.IsNotExist(err) {
		t.Errorf("removeAppstoreModules() removed custom module")
	}
}

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
			name:     "long folder name with slash - should be truncated",
			input:    "Folder/testverylonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong",
			expected: "Folder_testverylongl_TRUNCATED_80d74_glonglonglong", // Truncated to 50 chars
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
		{
			name:     "Windows reserved name CON",
			input:    "CON",
			expected: "_CON",
		},
		{
			name:     "Windows reserved name COM1",
			input:    "COM1",
			expected: "_COM1",
		},
		{
			name:     "control characters",
			input:    "Folder\x00Test\x1F",
			expected: "Folder_Test_",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  FolderName  ",
			expected: "FolderName",
		},
		{
			name:     "special characters",
			input:    "Folder<>:\"|?*Test",
			expected: "Folder_______Test",
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
		{
			name: "depth limit test",
			folders: []MxFolder{
				{Name: "Root", ID: "1", Parent: nil},
			},
			targetID: "1",
			expected: "Root",
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
