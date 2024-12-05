package main

import (
	"fmt"
	"os"
)

const (
	version = "v0.0.33"

	githubAPI = "https://api.github.com/repos/codeyarduk/mordecai/releases/latest"
)

var supportedFileTypes = []string{
	".jsx", ".tsx", ".json", ".html", ".css", ".md", ".yml", ".yaml",
	".scss", ".svelte", ".vue", ".py", ".go", ".c", ".rs", ".rb",
	".zig", ".php", ".ts", ".mts", ".cts", ".js", ".mjs", ".cjs",
}

var (
	siteUrl = "devwilson.dev"
)

//                          _                _
//  _ __ ___   ___  _ __ __| | ___  ___ __ _(_)
// | '_ ` _ \ / _ \| '__/ _` |/ _ \/ __/ _` | |
// | | | | | | (_) | | | (_| |  __/ (_| (_| | |
// |_| |_| |_|\___/|_|  \__,_|\___|\___\__,_|_|
//

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: mordecai <command>")
		fmt.Println("Run 'mordecai --help' for a list of available commands.")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "link":

		if len(os.Args) < 2 {
			fmt.Println("Usage: mordecai <commands>")
			os.Exit(1)
		}

		updateVersion()
		linkCommand()

	case "logout":
		logoutCommand()
	case "--help":
		helpCommand()
	case "--version":
		versionCommand()
	case "--installation-method":
		var installationMethod, err = installationMethodCommand()
		if err != nil {
			fmt.Println("Error checking installation method:", err)
			return
		}
		fmt.Printf("Mordecai was installed with %s\n", installationMethod)
	default:
		fmt.Printf("Unknown command %s\n", command)
		fmt.Println("Use 'mordecai --help' for usage information.")
		os.Exit(1)
	}
}

//            _                                                   _
//  ___ _   _| |__   ___ ___  _ __ ___  _ __ ___   __ _ _ __   __| |___
// / __| | | | '_ \ / __/ _ \| '_ ` _ \| '_ ` _ \ / _` | '_ \ / _` / __|
// \__ \ |_| | |_) | (_| (_) | | | | | | | | | | | (_| | | | | (_| \__ \
// |___/\__,_|_.__/ \___\___/|_| |_| |_|_| |_| |_|\__,_|_| |_|\__,_|___/
//

func linkCommand() {

	// Check if token is valid
	if tokenIsValid, err := checkIfTokenIsValid(); err != nil {
		fmt.Printf("Error checking token: %v\n", err)
		return
	} else if !tokenIsValid {
		authenticate()
		// This needs to be updated soon
	}

	var token, tokenErr = loadToken()

	if tokenErr != nil {
		fmt.Printf("Error getting token: %v\n", tokenErr)
	}

	// Get all repote spaces
	workspaceId, workspaceName, err := getWorkspaces(token)
	if err != nil {
		fmt.Printf("Error getting workspaces: %v\n", err)
		return

	}
	// Get name of the context
	repoName, repoId, err := linkRepo(token, workspaceId)

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

	err = showLoadingAnimation("Initialising repository...", func() error {
		var sendErr error
		repoId, sendErr = sendDataToServer(dirContent, token, workspaceId, repoName, repoId, false)
		return sendErr
	})

	if err != nil {
		fmt.Println("Error sending data to server.")
		fmt.Println(err)
		return
	}

	fmt.Printf("\033[1;32m✓ Syncing local repository \033[1;36m%s\033[1;32m to remote space \033[1;36m%s\033[0m\n", repoName, workspaceName)
	fmt.Println("\033[1;33m⚠ ALERT: Please leave this open while programming\033[0m")

	// Add a watcher to the directory
	err = watchDirectory(currentDir, workspaceId, repoName, repoId, token)
	if err != nil {
		fmt.Printf("Error setting up directory watcher: %v\n", err)
		return
	}

}

func logoutCommand() {
	token, tokenErr := loadToken()

	if tokenErr != nil {
		fmt.Printf("Error loading token: %v\n", tokenErr)
		return
	}

	err := deleteToken()
	if err != nil {
		fmt.Printf("Error deleting token: %v\n", err)
		return
	}
	if len(token) > 0 {
		fmt.Println("Successfully logged out!")
		return
	}
	fmt.Println("No active session found.")
}

func versionCommand() {
	fmt.Printf("mordecai version %s\n", version)
}

func helpCommand() {
	fmt.Println("Mordecai CLI Usage:")
	fmt.Println("  mordecai link                   - Link your codebase with Mordecai")
	fmt.Println("  mordecai logout                 - Logout of your Mordecai account")
	fmt.Println("  mordecai --help                 - Display this help message")
	fmt.Println("  mordecai --version              - Display the version of Mordecai you have installed")
	fmt.Println("  mordecai --installation-method  - Display the method you used to install mordecai")
}
