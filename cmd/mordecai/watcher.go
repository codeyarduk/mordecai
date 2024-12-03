package main

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//                _       _           _ _               _
// __      ____ _| |_ ___| |__     __| (_)_ __ ___  ___| |_ ___  _ __ _   _
// \ \ /\ / / _` | __/ __| '_ \   / _` | | '__/ _ \/ __| __/ _ \| '__| | | |
//  \ V  V / (_| | || (__| | | | | (_| | | | |  __/ (__| || (_) | |  | |_| |
//   \_/\_/ \__,_|\__\___|_| |_|  \__,_|_|_|  \___|\___|\__\___/|_|   \__, |
//                                                                    |___/

// FIX THE MEMORY LEAKS

func watchDirectory(directoryPath, workspaceId, repoName, repoId, token string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %v", err)
	}

	defer watcher.Close()

	filesToUpdate := make([]FileContent, 0)
	var timeoutTimer *time.Timer

	// Define directories to ignore
	ignorePatterns, err := readGitignore(directoryPath)
	if err != nil {
		return fmt.Errorf("error reading .gitignore: %v", err)
	}

	err = filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Check if the directory should be ignored
			if shouldIgnore(path, ignorePatterns) {
				return filepath.SkipDir
			}

			// fmt.Printf("Watching directory: %s\n", path)
			_ = watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error setting up recursive watch: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher channel closed")
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) != 0 {
				// Check if file path must be ignored
				filePath := event.Name
				if shouldIgnore(filePath, ignorePatterns) {
					continue
				}

				// Check if the file extension is allowed
				fileExtension := filepath.Ext(filePath)
				if !contains(supportedFileTypes, fileExtension) {
					continue
				}

				// Check if file is already in filesToUpdate
				fileRepeated := false
				for _, file := range filesToUpdate {
					if file.FilePath == filePath {
						fileRepeated = true
						break
					}
				}

				if !fileRepeated {
					content, err := readFile(filePath)
					if err != nil {
						fmt.Printf("Error reading file %s: %v\n", filePath, err)
						continue
					}

					filesToUpdate = append(filesToUpdate, FileContent{
						FilePath:      filePath,
						FileExtension: fileExtension,
						DataChunks:    content,
					})
				}

				// If a new directory is created, add it to the watcher
				if info, err := os.Stat(filePath); err == nil && info.IsDir() {
					if !shouldIgnore(filePath, ignorePatterns) {
						err = watcher.Add(filePath)
						if err != nil {
							fmt.Printf("Error watching new directory %s: %v\n", filePath, err)
						} else {
							fmt.Printf("New directory added to watch: %s\n", filePath)
						}
					}
				}

				if timeoutTimer != nil {
					timeoutTimer.Stop()
				}

				timeoutTimer = time.AfterFunc(5*time.Second, func() {
					processUpdatedFiles(filesToUpdate, token, workspaceId, repoId, repoName)
					filesToUpdate = make([]FileContent, 0)
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher error channel closed")
			}
			fmt.Println("error:", err)
		}
	}
}

func readGitignore(dirPath string) ([]string, error) {
	gitignorePath := filepath.Join(dirPath, ".gitignore")
	ignorePatterns := []string{".git", "node_modules", "package-lock.json"}

	if _, err := os.Stat(gitignorePath); err == nil {
		file, err := os.Open(gitignorePath)
		if err != nil {
			return nil, fmt.Errorf("error opening .gitignore: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			pattern := strings.TrimSpace(scanner.Text())
			if pattern != "" && !strings.HasPrefix(pattern, "#") {
				ignorePatterns = append(ignorePatterns, pattern)
			}
		}
	}

	return ignorePatterns, nil
}

func shouldIgnore(path string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
		if strings.HasPrefix(pattern, "/") {
			if strings.HasPrefix(path, filepath.Clean(pattern)) {
				return true
			}
		} else {
			if strings.Contains(path, string(filepath.Separator)+pattern+string(filepath.Separator)) {
				return true
			}
		}
	}
	return false
}

// Helper function to check if a directory should be ignored
func containsDir(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to check if a file is in an ignored directory
func isInIgnoredDir(filePath string, ignoreDirs []string) bool {
	parts := strings.Split(filePath, string(os.PathSeparator))
	for _, part := range parts {
		if containsDir(ignoreDirs, part) {
			return true
		}
	}
	return false
}

// Helper function to check if a slice contains a string

func processUpdatedFiles(filesToUpdate []FileContent, token, workspaceId string, repoId string, repoName string) {
	err := showLoadingAnimation("Updating files...", func() error {
		_, err := sendDataToServer(filesToUpdate, token, workspaceId, repoName, repoId, true)
		return err
	})

	if err != nil {
		fmt.Printf("Error processing files: %v\n", err)
		return
	}
}
