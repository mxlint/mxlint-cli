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

		script := `mxlint.io.readfile("test.txt")`
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

		script := `mxlint.io.readfile("` + testFilePath + `")`
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
			mxlint.io.readfile("nonexistent.txt");
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

	t.Run("path traversal with .. is blocked", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.readfile("../../../etc/passwd");
			"no error";
		} catch (e) {
			e.message.includes("outside working directory") ? "blocked" : "other error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != "blocked" {
			t.Errorf("Expected path traversal to be blocked, got: %s", result.String())
		}
	})

	t.Run("absolute path outside working directory is blocked", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.readfile("/etc/passwd");
			"no error";
		} catch (e) {
			e.message.includes("outside working directory") ? "blocked" : "other error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != "blocked" {
			t.Errorf("Expected absolute path outside working dir to be blocked, got: %s", result.String())
		}
	})

	t.Run("readfile without argument throws error", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.readfile();
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

		script := `mxlint.io.readfile("subdir/subfile.txt")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != subFileContent {
			t.Errorf("Expected %q, got %q", subFileContent, result.String())
		}
	})
}

func TestSetupJavascriptVM_MxlintListdir(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create some test files and directories
	err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "subdir", "nested.txt"), []byte("nested"), 0644)
	if err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	t.Run("list directory with relative path", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `JSON.stringify(mxlint.io.listdir(".").sort())`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		expected := `["file1.txt","file2.txt","subdir"]`
		if result.String() != expected {
			t.Errorf("Expected %q, got %q", expected, result.String())
		}
	})

	t.Run("list directory with absolute path", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `JSON.stringify(mxlint.io.listdir("` + tempDir + `").sort())`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		expected := `["file1.txt","file2.txt","subdir"]`
		if result.String() != expected {
			t.Errorf("Expected %q, got %q", expected, result.String())
		}
	})

	t.Run("list subdirectory", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `JSON.stringify(mxlint.io.listdir("subdir"))`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		expected := `["nested.txt"]`
		if result.String() != expected {
			t.Errorf("Expected %q, got %q", expected, result.String())
		}
	})

	t.Run("list nonexistent directory throws error", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.listdir("nonexistent");
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
			t.Error("Expected an error when listing nonexistent directory")
		}
	})

	t.Run("path traversal with .. is blocked", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.listdir("../../../etc");
			"no error";
		} catch (e) {
			e.message.includes("outside working directory") ? "blocked" : "other error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != "blocked" {
			t.Errorf("Expected path traversal to be blocked, got: %s", result.String())
		}
	})

	t.Run("absolute path outside working directory is blocked", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.listdir("/etc");
			"no error";
		} catch (e) {
			e.message.includes("outside working directory") ? "blocked" : "other error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != "blocked" {
			t.Errorf("Expected absolute path outside working dir to be blocked, got: %s", result.String())
		}
	})

	t.Run("listdir without argument throws error", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.listdir();
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
			t.Error("Expected an error when calling listdir without argument")
		}
	})

	t.Run("list empty directory", func(t *testing.T) {
		// Create an empty subdirectory
		emptyDir := filepath.Join(tempDir, "empty")
		err := os.Mkdir(emptyDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		vm := setupJavascriptVM(tempDir)

		script := `JSON.stringify(mxlint.io.listdir("empty"))`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		expected := `[]`
		if result.String() != expected {
			t.Errorf("Expected %q, got %q", expected, result.String())
		}
	})

	t.Run("listdir on file throws error", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.listdir("file1.txt");
			"no error";
		} catch (e) {
			e.message.includes("not a directory") ? "not a directory" : "error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() == "no error" {
			t.Error("Expected an error when calling listdir on a file")
		}
	})
}

func TestSetupJavascriptVM_MxlintIsdir(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create some test files and directories
	err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	t.Run("isdir returns true for directory with relative path", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `mxlint.io.isdir("subdir")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.ToBoolean() != true {
			t.Errorf("Expected true for directory, got %v", result.ToBoolean())
		}
	})

	t.Run("isdir returns true for directory with absolute path", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `mxlint.io.isdir("` + filepath.Join(tempDir, "subdir") + `")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.ToBoolean() != true {
			t.Errorf("Expected true for directory, got %v", result.ToBoolean())
		}
	})

	t.Run("isdir returns false for file", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `mxlint.io.isdir("file.txt")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.ToBoolean() != false {
			t.Errorf("Expected false for file, got %v", result.ToBoolean())
		}
	})

	t.Run("isdir returns false for nonexistent path", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `mxlint.io.isdir("nonexistent")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.ToBoolean() != false {
			t.Errorf("Expected false for nonexistent path, got %v", result.ToBoolean())
		}
	})

	t.Run("isdir returns true for current directory", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `mxlint.io.isdir(".")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.ToBoolean() != true {
			t.Errorf("Expected true for current directory, got %v", result.ToBoolean())
		}
	})

	t.Run("path traversal with .. is blocked", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.isdir("../../../etc");
			"no error";
		} catch (e) {
			e.message.includes("outside working directory") ? "blocked" : "other error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != "blocked" {
			t.Errorf("Expected path traversal to be blocked, got: %s", result.String())
		}
	})

	t.Run("absolute path outside working directory is blocked", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.isdir("/etc");
			"no error";
		} catch (e) {
			e.message.includes("outside working directory") ? "blocked" : "other error: " + e.message;
		}
		`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.String() != "blocked" {
			t.Errorf("Expected absolute path outside working dir to be blocked, got: %s", result.String())
		}
	})

	t.Run("isdir without argument throws error", func(t *testing.T) {
		vm := setupJavascriptVM(tempDir)

		script := `
		try {
			mxlint.io.isdir();
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
			t.Error("Expected an error when calling isdir without argument")
		}
	})

	t.Run("isdir with nested directory path", func(t *testing.T) {
		// Create a nested directory
		nestedDir := filepath.Join(tempDir, "subdir", "nested")
		err := os.Mkdir(nestedDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create nested directory: %v", err)
		}

		vm := setupJavascriptVM(tempDir)

		script := `mxlint.io.isdir("subdir/nested")`
		result, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		if result.ToBoolean() != true {
			t.Errorf("Expected true for nested directory, got %v", result.ToBoolean())
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

	// Check that mxlint.io is an object
	script = `typeof mxlint.io`
	result, err = vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "object" {
		t.Errorf("Expected mxlint.io to be an object, got %q", result.String())
	}

	// Check that mxlint.io.readfile is a function
	script = `typeof mxlint.io.readfile`
	result, err = vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "function" {
		t.Errorf("Expected mxlint.io.readfile to be a function, got %q", result.String())
	}

	// Check that mxlint.io.listdir is a function
	script = `typeof mxlint.io.listdir`
	result, err = vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "function" {
		t.Errorf("Expected mxlint.io.listdir to be a function, got %q", result.String())
	}

	// Check that mxlint.io.isdir is a function
	script = `typeof mxlint.io.isdir`
	result, err = vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "function" {
		t.Errorf("Expected mxlint.io.isdir to be a function, got %q", result.String())
	}
}
