package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

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
			fmt.Println(token)
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

//            _                                                   _
//  ___ _   _| |__   ___ ___  _ __ ___  _ __ ___   __ _ _ __   __| |___
// / __| | | | '_ \ / __/ _ \| '_ ` _ \| '_ ` _ \ / _` | '_ \ / _` / __|
// \__ \ |_| | |_) | (_| (_) | | | | | | | | | | | (_| | | | | (_| \__ \
// |___/\__,_|_.__/ \___\___/|_| |_| |_|_| |_| |_|\__,_|_| |_|\__,_|___/
//

func linkCommand() {
	openBrowser("https://devwilson.dev/login?cli=true?port=8300")
	startLocalServer(8300)
	fmt.Println("Linking your codebase with Mordecai...")
}

func logoutCommand() {
	var token, error = loadToken()
	fmt.Println(token)
	fmt.Println(error)
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
