package main

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"os"
	"path/filepath"
	"strings"
)

//
//   __ _ _
//  / _(_) | ___
// | |_| | |/ _ \
// |  _| | |  __/
// |_| |_|_|\___|
//
//  _ __ ___   __ _ _ __   __ _  __ _  ___ _ __ ___   ___ _ __ | |_
// | '_ ` _ \ / _` | '_ \ / _` |/ _` |/ _ \ '_ ` _ \ / _ \ '_ \| __|
// | | | | | | (_| | | | | (_| | (_| |  __/ | | | | |  __/ | | | |_
// |_| |_| |_|\__,_|_| |_|\__,_|\__, |\___|_| |_| |_|\___|_| |_|\__|
//                              |___/

func readFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v", filePath, err)
		return "", err
	}
	return string(content), nil
}

func readDir(dirPath string) ([]string, error) {
	var files []string

	// Create patterns list
	ps := []gitignore.Pattern{
		gitignore.ParsePattern(".git", nil),
		gitignore.ParsePattern("node_modules", nil),
		gitignore.ParsePattern("package-lock.json", nil),
	}

	// Read .gitignore file if it exists
	gitignorePath := filepath.Join(dirPath, ".gitignore")
	if data, err := os.ReadFile(gitignorePath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				ps = append(ps, gitignore.ParsePattern(line, nil))
			}
		}
	}

	// Create matcher
	matcher := gitignore.NewMatcher(ps)

	// Walk directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for gitignore matching
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Skip if matched by gitignore
		if matcher.Match(strings.Split(relPath, string(filepath.Separator)), info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Add file if it's not a directory and has supported extension
		if !info.IsDir() {
			ext := filepath.Ext(path)
			for _, allowedExt := range supportedFileTypes {
				if ext == allowedExt {
					files = append(files, path)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return files, nil
}

type FileContent struct {
	FilePath      string `json:"file_path"`
	FileExtension string `json:"file_extension"`
	DataChunks    string `json:"data_chunks"`
}

func getFileContents(files []string) ([]FileContent, error) {
	var fileContents []FileContent

	for _, filePath := range files {
		// Check if it's a regular file
		info, err := os.Stat(filePath)
		if err != nil {
			return nil, fmt.Errorf("error getting file info for %s: %v", filePath, err)
			// Should a continue be here?
		}
		if !info.Mode().IsRegular() {
			continue // Skip non-regular files
		}

		// Get file extension
		ext := filepath.Ext(filePath)

		// Check if it's an allowed file type (you can modify this list as needed)

		if !contains(supportedFileTypes, ext) {
			continue
		}

		// Read file content
		content, err := readFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
			// Should a continue be here?

		}

		fileContents = append(fileContents, FileContent{
			FilePath:      filePath,
			DataChunks:    content,
			FileExtension: ext,
		})
	}

	return fileContents, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
