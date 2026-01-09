package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupJavascriptVM_MxlintReadfile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test file to read
	testContent := "Hello, mxlint!"
	testFilePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("read file with relative path", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `mxlint.readfile("test.txt")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != testContent {
			t.Errorf("Expected %q, got %q", testContent, result.String())
		}
	})

	t.Run("read file with absolute path", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `mxlint.readfile("` + testFilePath + `")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != testContent {
			t.Errorf("Expected %q, got %q", testContent, result.String())
		}
	})

	t.Run("read nonexistent file throws error", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.readfile("nonexistent.txt");
			"no error";
		} catch (e) {
			"error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() == "no error" {
			t.Error("Expected an error when reading nonexistent file")
		}
	})

	t.Run("readfile without argument throws error", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.readfile();
			"no error";
		} catch (e) {
			"error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() == "no error" {
			t.Error("Expected an error when calling readfile without argument")
		}
	})

	t.Run("read file in subdirectory", func(t *testing.T) {
		// Create a subdirectory with a test file
		subDir := filepath.Join(tempDir, "subdir")
		err := os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		subFileContent := "Content in subdirectory"
		subFilePath := filepath.Join(subDir, "subfile.txt")
		err = os.WriteFile(subFilePath, []byte(subFileContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create subfile: %v", err)
		}

		vm := setupJavascriptVM(tempDir)

		script := `mxlint.readfile("subdir/subfile.txt")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != subFileContent {
			t.Errorf("Expected %q, got %q", subFileContent, result.String())
		}
	})
}

func TestMxlintObjectAvailable(t *testing.T) {
	vm := setupJavascriptVM(".")

	// Check that mxlint object is available
	script := `typeof mxlint`
	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "object" {
		t.Errorf("Expected mxlint to be an object, got %q", result.String())
	}

	// Check that mxlint.readfile is a function
	script = `typeof mxlint.readfile`
	result, err = vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "function" {
		t.Errorf("Expected mxlint.readfile to be a function, got %q", result.String())
	}
}
