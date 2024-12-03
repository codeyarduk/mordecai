package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile(t *testing.T) {
	// Create a temporary test file
	content := "test content"
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test reading the file
	result, err := readFile(tmpFile.Name())
	if err != nil {
		t.Errorf("readFile() error = %v", err)
	}
	if result != content {
		t.Errorf("readFile() = %v, want %v", result, content)
	}
}

func TestReadDir(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{
		"test1.go",
		"test2.js",
		"test3.txt",
	}

	for _, fname := range testFiles {
		tmpfn := filepath.Join(tmpDir, fname)
		if err := os.WriteFile(tmpfn, []byte("test"), 0666); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test reading the directory
	files, err := readDir(tmpDir)
	if err != nil {
		t.Errorf("readDir() error = %v", err)
	}

	// Verify only supported file types are returned
	for _, file := range files {
		ext := filepath.Ext(file)
		if !contains(supportedFileTypes, ext) {
			t.Errorf("readDir() returned unsupported file type: %v", ext)
		}
	}
}

func TestGetFileContents(t *testing.T) {
	// Create temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.go")
	content := "package main\n\nfunc main() {}"
	if err := os.WriteFile(testFile, []byte(content), 0666); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test getting file contents
	files := []string{testFile}
	contents, err := getFileContents(files)
	if err != nil {
		t.Errorf("getFileContents() error = %v", err)
	}

	if len(contents) != 1 {
		t.Errorf("getFileContents() returned %v files, want 1", len(contents))
	}

	if contents[0].FileExtension != ".go" {
		t.Errorf("getFileContents() extension = %v, want .go", contents[0].FileExtension)
	}

	if contents[0].DataChunks != content {
		t.Errorf("getFileContents() content = %v, want %v", contents[0].DataChunks, content)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
	}

	for _, tt := range tests {
		result := contains(tt.slice, tt.item)
		if result != tt.expected {
			t.Errorf("contains(%v, %v) = %v, want %v",
				tt.slice, tt.item, result, tt.expected)
		}
	}
}
