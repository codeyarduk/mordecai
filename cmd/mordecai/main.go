package main

import (
	"fmt"
	// "github.com/fsnotify/fsnotify"
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/AlecAivazis/survey/v2"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//                          _                _
//  _ __ ___   ___  _ __ __| | ___  ___ __ _(_)
// | '_ ` _ \ / _ \| '__/ _` |/ _ \/ __/ _` | |
// | | | | | | (_) | | | (_| |  __/ (_| (_| | |
// |_| |_| |_|\___/|_|  \__,_|\___|\___\__,_|_|
//

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mordecai <commands>")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "link":
		linkCommand()
	case "logout":
		logoutCommand()
	case "--help":
		helpCommand()
	default:
		fmt.Printf("Unknown command %s\n", command)
		fmt.Println("Use 'mordecai --help' for usage information.")
		os.Exit(1)
	}
}

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
	ignorePatterns := []string{".git"}

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

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file/directory should be ignored
		for _, pattern := range ignorePatterns {
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
		}

		if !info.IsDir() {
			files = append(files, path)
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
		}
		if !info.Mode().IsRegular() {
			continue // Skip non-regular files
		}

		// Get file extension
		ext := filepath.Ext(filePath)

		// Check if it's an allowed file type (you can modify this list as needed)
		allowedExtensions := []string{".go", ".js", ".ts", ".py", ".html", ".css", ".json", ".rb", ".md"}
		if !contains(allowedExtensions, ext) {
			continue
		}

		// Read file content
		content, err := readFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
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

//  _
// | |__  _ __ _____      _____  ___ _ __
// | '_ \| '__/ _ \ \ /\ / / __|/ _ \ '__|
// | |_) | | | (_) \ V  V /\__ \  __/ |
// |_.__/|_|  \___/ \_/\_/ |___/\___|_|
//              _   _                _   _           _   _
//   __ _ _   _| |_| |__   ___ _ __ | |_(_) ___ __ _| |_(_) ___  _ __
//  / _` | | | | __| '_ \ / _ \ '_ \| __| |/ __/ _` | __| |/ _ \| '_ \
// | (_| | |_| | |_| | | |  __/ | | | |_| | (_| (_| | |_| | (_) | | | |
//  \__,_|\__,_|\__|_| |_|\___|_| |_|\__|_|\___\__,_|\__|_|\___/|_| |_|
//

func authenticate() (string, error) {

	token, err := loadToken()

	if err != nil {
		return "", fmt.Errorf("failed to load token: %w", err)
	}

	if len(token) > 0 {
		fmt.Println("You're already signed in HURRAYYYY")
		// This is where you will ping the workspaces to see if the token is valid
	}

	port := 8300
	authURL := fmt.Sprintf("%shttps://api.devwilson.dev/auth/cli?port=%d", os.Getenv("SITE_URL"), port)
	// Start local server in a goroutine
	tokenChan := make(chan string, 1)
	errChan := make(chan error, 1)
	go func() {
		token, err := startLocalServer(port)
		if err != nil {
			errChan <- err
		} else {
			tokenChan <- token
		}
	}()

	// Open browser
	if err := openBrowser(authURL); err != nil {
		return "", fmt.Errorf("failed to open browser: %w", err)
	}

	// Wait for token or error with a timeout
	select {
	case token := <-tokenChan:
		return token, nil
	case err := <-errChan:
		return "", err
	case <-time.After(2 * time.Minute):
		return "", fmt.Errorf("authentication timed out")
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}

func startLocalServer(callbackPort int) (string, error) {
	tokenChan := make(chan string, 1)
	errChan := make(chan error, 1)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		parsedURL, err := url.Parse(r.URL.String())
		if err != nil {
			errChan <- fmt.Errorf("failed to parse URL: %w", err)
			return
		}

		token := parsedURL.Query().Get("token")
		if token != "" {
			saveToken(token)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "<h1>Authentication successful! You can close this window.</h1>")
			tokenChan <- token
		} else {
			errChan <- fmt.Errorf("no token received")
		}
	})

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", callbackPort), nil); err != nil {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case token := <-tokenChan:
		return token, nil
	case err := <-errChan:
		return "", err
	}
}

//  _        _
// | |_ ___ | | _____ _ __  ___
// | __/ _ \| |/ / _ \ '_ \/ __|
// | || (_) |   <  __/ | | \__ \
//  \__\___/|_|\_\___|_| |_|___/
//

func checkIfTokenIsValid() (bool, error) {
	var token, err = loadToken()
	if err != nil {
		return false, err
	}

	if len(token) > 0 {
		return true, nil
	}

	return false, nil
}

func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	mordecaiPath := filepath.Join(homeDir, ".mordecai")
	if err := os.MkdirAll(mordecaiPath, 0700); err != nil {
		return "", fmt.Errorf("failed to create .mordecai directory: %w", err)
	}
	return filepath.Join(mordecaiPath, ".mordecai_token"), nil
}

func saveToken(token string) error {
	filePath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, []byte(token), 0600)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	return nil
}

func loadToken() (string, error) {
	filePath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No token file found, but not an error
		}
		return "", fmt.Errorf("failed to load token: %w", err)
	}
	return string(data), nil
}

func deleteToken() error {
	filePath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

//                  _           _
//  _ __  _ __ ___ (_) ___  ___| |_ ___
// | '_ \| '__/ _ \| |/ _ \/ __| __/ __|
// | |_) | | | (_) | |  __/ (__| |_\__ \
// | .__/|_|  \___// |\___|\___|\__|___/
// |_|           |__/
//

func getWorkspaces(token string) (string, error) {
	fmt.Println("Fetching available workspaces...")

	const siteUrl string = "https://api.devwilson.dev"
	endpointURL := fmt.Sprintf("%s/cli/workspaces", siteUrl)

	// Create the request body
	postData := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	// Marshal the postData into JSON
	jsonData, err := json.Marshal(postData)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get workspaces. Status: %s", resp.Status)
	}

	// Read and parse the response body
	var workspaces []struct {
		WorkspaceID   string `json:"workspaceId"`
		WorkspaceName string `json:"workspaceName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	// Create choices for user selection
	var choices []string
	for _, workspace := range workspaces {
		choices = append(choices, workspace.WorkspaceName)
	}

	// Prompt user to select a workspace
	prompt := &survey.Select{
		Message: "Select the project that you want to link this directory to:",
		Options: choices,
	}

	var selectedName string
	if err := survey.AskOne(prompt, &selectedName); err != nil {
		return "", fmt.Errorf("error during workspace selection: %v", err)
	}

	// Find the selected workspace ID
	for _, workspace := range workspaces {
		if workspace.WorkspaceName == selectedName {
			return workspace.WorkspaceID, nil
		}
	}

	return "", fmt.Errorf("selected workspace not found")
}

//            _                                                   _
//  ___ _   _| |__   ___ ___  _ __ ___  _ __ ___   __ _ _ __   __| |___
// / __| | | | '_ \ / __/ _ \| '_ ` _ \| '_ ` _ \ / _` | '_ \ / _` / __|
// \__ \ |_| | |_) | (_| (_) | | | | | | | | | | | (_| | | | | (_| \__ \
// |___/\__,_|_.__/ \___\___/|_| |_| |_|_| |_| |_|\__,_|_| |_|\__,_|___/
//

func linkCommand() {

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	var dir, dirErr = readDir(currentDir)
	if dirErr != nil {
		fmt.Printf("Error reading current directory: %v\n", err)
	}

	var dirContent, dirContentErr = getFileContents(dir)
	if dirContentErr != nil {
		fmt.Printf("Error reading current directory: %v\n", err)
	}

	if tokenIsValid, err := checkIfTokenIsValid(); err != nil {
		fmt.Printf("Error checking token: %v\n", err)
		return
	} else if !tokenIsValid {
		authenticate()

		// Perform the action when the token is valid
		// Add further code here as needed
	}

	var token, tokenErr = loadToken()

	if tokenErr != nil {
		fmt.Println("Error getting token: %v\n", tokenErr)
	}

	workspaceID, err := getWorkspaces(token)
	if err != nil {
		fmt.Printf("Error getting workspaces: %v\n", err)
		return
	}

	fmt.Println("Linking your codebase with Mordecai...")
	const siteUrl string = "https://api.devwilson.dev"
	endpointURL := fmt.Sprintf("%s/cli/chunk", siteUrl)

	postData := struct {
		Files       []FileContent `json:"files"`
		Token       string        `json:"token"`
		WorkspaceId string        `json:"workspaceId,omitempty"`
	}{
		Files:       dirContent,
		Token:       token,
		WorkspaceId: workspaceID, // Use the workspaceID you obtained earlier
	}
	// Marshal the data to JSON
	jsonData, err := json.Marshal(postData)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// Send the POST request
	req, err := http.NewRequest("POST", endpointURL, bytes.NewReader(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Handle the response status code and body as needed

}

func logoutCommand() {
	err := deleteToken()
	if err != nil {
		fmt.Printf("Error deleting token: %v\n", err)
		return
	}
	fmt.Println("Successfully logged out!")
}

func helpCommand() {
	fmt.Println("Mordecai CLI Usage:")
	fmt.Println("  mordecai link     - Link your codebase with Mordecai")
	fmt.Println("  mordecai --help   - Display this help message")
}
