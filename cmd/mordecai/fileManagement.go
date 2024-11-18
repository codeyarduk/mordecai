package main

import (
	"bufio"
	"fmt"
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

	// Read .gitignore file
	gitignorePath := filepath.Join(dirPath, ".gitignore")
	ignorePatterns := []string{".git", "node_modules", "package-lock.json"}

	if _, err := os.Stat(gitignorePath); err == nil {
		file, err := os.Open(gitignorePath)
		if err != nil {
			return nil, fmt.Errorf("error opening .gitignore: %v", err)
		}
		defer file.Close()
		// What is this doing?
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			pattern := strings.TrimSpace(scanner.Text())
			if pattern != "" && !strings.HasPrefix(pattern, "#") {
				ignorePatterns = append(ignorePatterns, pattern)
			}
		}
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file/directory should be ignored
		for _, pattern := range ignorePatterns {
			// Check if the pattern matches the base name
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if the pattern is anywhere in the path
			if strings.HasPrefix(pattern, "/") {
				// Absolute path pattern
				if strings.HasPrefix(path, filepath.Clean(pattern)) {
					return filepath.SkipDir
				}
			} else {
				// Relative path pattern
				if strings.Contains(path, string(filepath.Separator)+pattern+string(filepath.Separator)) {
					return filepath.SkipDir
				}
			}
		}

		if !info.IsDir() {
			// Check if the file extension is in the allowed list
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
