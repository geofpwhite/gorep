package main

import (
	"os"
	"regexp"
	"testing"
)

func TestInvalidArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldPanic bool
	}{
		{
			name:        "NoArguments",
			args:        []string{"gorep"},
			shouldPanic: true,
		},
		{
			name:        "InvalidRegex",
			args:        []string{"gorep", "[invalid", "valid", "test.txt"},
			shouldPanic: true,
		},
		{
			name:        "InvalidFlag",
			args:        []string{"gorep", "-invalid", "test", "replacement"},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic for test case: %s", tt.name)
					}
				}()
			}
			Configure() // Call the main function directly
		})
	}
}

func TestValidArgs(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("can't get pwd")
	}
	os.Args = []string{"gorep", "test", "this is a test"}
	c := newConfig(regexp.MustCompile("const[a-z]{3}"), true, ".", t.TempDir()+"/result", nil)
	if c.Main() != 0 {
		t.Fail()
	}
	t.Chdir(pwd)
	c.file = "main_test.go"
	c.outputPath = ""
	if c.Main() != 0 {
		t.Fail()
	}
}

func TestWalk(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", pwd)
	}
	tempDir := t.TempDir()
	files, err := walk(tempDir)
	if len(files) != 0 || err != nil {
		t.Errorf("Expected no files, got %d", len(files))
	}

	tempDir = t.TempDir()
	filePath := tempDir + "/file.txt"
	err = os.WriteFile(filePath, []byte("content"), 0o644)
	t.Chdir(pwd)
	if err != nil {
		t.Errorf("%v", err)
	}
	files, err = walk(tempDir)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(files) != 1 || files[0][0] != "file.txt" {
		t.Errorf("Expected one file named 'file.txt', got %v", files)
	}

	tempDir = t.TempDir()
	t.Chdir(pwd)
	nestedDir := tempDir + "/nested"
	os.Mkdir(nestedDir, 0o755)
	filePath = nestedDir + "/file.txt"
	os.WriteFile(filePath, []byte("content"), 0o644)
	files, err = walk(tempDir)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(files) != 1 || files[0][0] != "file.txt" {
		t.Errorf("Expected one file named 'file.txt' in nested directory, got %v", files)
	}
}
