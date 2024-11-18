package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

const (
	version = "v0.0.25"

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

//  ___  ___ _ ____   _____ _ __
// / __|/ _ \ '__\ \ / / _ \ '__|
// \__ \  __/ |   \ V /  __/ |
// |___/\___|_|    \_/ \___|_|
//                             _   _
//   ___  _ __   ___ _ __ __ _| |_(_) ___  _ __  ___
//  / _ \| '_ \ / _ \ '__/ _` | __| |/ _ \| '_ \/ __|
// | (_) | |_) |  __/ | | (_| | |_| | (_) | | | \__ \
//  \___/| .__/ \___|_|  \__,_|\__|_|\___/|_| |_|___/
//       |_|
//

func getWorkspaces(token string) (string, string, error) {
	fmt.Println("Fetching available workspaces...")
	endpointURL := fmt.Sprintf("https://api.%s/cli/spaces", siteUrl)

	// Create the request body
	postData := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	// Marshal the postData into JSON
	jsonData, err := json.Marshal(postData)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to get workspaces. Status: %s", resp.Status)
	}

	// Read and parse the response body
	var workspaces []struct {
		WorkspaceID   string `json:"spaceId"`
		WorkspaceName string `json:"spaceName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
		return "", "", fmt.Errorf("error decoding response: %v", err)
	}

	// Clear the screen and move cursor to top before showing workspace selection
	fmt.Print("\033[2J")
	fmt.Print("\033[H")

	// Create a new workspace model with the enhanced styling
	m := newWorkspaceModel(workspaces)

	// Run the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	selectedModel, err := p.Run()
	if err != nil {
		return "", "", fmt.Errorf("error running workspace selection: %v", err)
	}

	// Get the selected workspace
	selectedWorkspace := selectedModel.(workspaceModel)
	selectedId := selectedWorkspace.selectedId
	selectedName := selectedWorkspace.selectedName
	// Clear the screen again
	fmt.Print("\033[2J")
	fmt.Print("\033[H")

	// Display the syncing message

	return selectedId, selectedName, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#7D56F4")).
				SetString("► ")
)

// DUPLICATED TYPE
type workspace struct {
	id   string
	name string
}

func (w workspace) Title() string       { return w.name }
func (w workspace) Description() string { return "" } // Return an empty string
func (w workspace) FilterValue() string { return w.name }

type workspaceModel struct {
	list         list.Model
	selectedId   string
	selectedName string
}

func (m workspaceModel) Init() tea.Cmd {
	return nil
}

func (m workspaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(workspace)
			if ok {
				m.selectedId = i.id
				m.selectedName = i.name
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m workspaceModel) View() string {
	return docStyle.Render(m.list.View())
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func newWorkspaceModel(workspaces []struct {
	WorkspaceID   string `json:"spaceId"`
	WorkspaceName string `json:"spaceName"`
}) workspaceModel {
	items := make([]list.Item, len(workspaces))
	for i, w := range workspaces {
		items[i] = workspace{id: w.WorkspaceID, name: w.WorkspaceName}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false // Hide the description
	delegate.Styles.NormalTitle = itemStyle
	delegate.Styles.SelectedTitle = selectedItemStyle.Inline(true).
		Foreground(lipgloss.Color("#7D56F4"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Select a Space"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4)

	return workspaceModel{list: l}
}

// Get current repository name

func getRepoName() (string, error) {
	// Try to get the remote URL
	remoteURL, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err == nil && len(remoteURL) > 0 {
		// Extract repo name from remote URL
		return extractRepoNameFromURL(string(remoteURL)), nil
	}

	// If no remote, get the current directory name
	dir, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	return filepath.Base(dir), nil
}

func extractRepoNameFromURL(url string) string {
	// Remove newline and trailing .git if present
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")

	// Split the URL and get the last part
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func linkRepo(token string, workspaceId string) (string, string, error) {
	endpointURL := fmt.Sprintf("https://api.%s/cli/space-repositories", siteUrl)
	currentRepoName, err := getRepoName()

	if err != nil {
		return "", "", fmt.Errorf("error getting the current repo name: %v", err)
	}

	// Create the request body
	postData := struct {
		Token       string `json:"token"`
		WorkspaceId string `json:"spaceId"`
	}{
		Token:       token,
		WorkspaceId: workspaceId,
	}

	// Marshal the postData into JSON
	jsonData, err := json.Marshal(postData)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to get workspaces. Status: %s", resp.Status)
	}

	// Read and parse the response body
	var repos []struct {
		RepoID   string `json:"contextId"`
		RepoName string `json:"contextName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return "", "", fmt.Errorf("error decoding response: %v", err)
	}

	// Checks if current repo has been previously linked
	for _, repo := range repos {
		if repo.RepoName == currentRepoName {
			return currentRepoName, repo.RepoID, nil
		}
	}

	selectedRepoName, selectedRepoId := currentRepoName, ""

	return selectedRepoName, selectedRepoId, nil
}

func sendDataToServer(files []FileContent, token string, workspaceId string, repoName string, repoId string, update bool) error {

	endpointURL := fmt.Sprintf("https://api.%s/cli/chunk", siteUrl)

	postData := struct {
		Files       []FileContent `json:"files"`
		Token       string        `json:"token"`
		ContextId   string        `json:"contextId,omitempty"`
		ContextName string        `json:"contextName"`
		WorkspaceId string        `json:"spaceId,omitempty"`
		Update      bool          `json:"update"`
	}{
		Files:       files,
		Token:       token,
		ContextId:   repoId,
		ContextName: repoName,
		WorkspaceId: workspaceId, // Use the workspaceID you obtained earlier
		Update:      update,
	}

	jsonData, err := json.Marshal(postData)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return err
	}

	// Send the POST request
	req, err := http.NewRequest("POST", endpointURL, bytes.NewReader(jsonData))

	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	// Handle the response status code and body as needed
	return nil
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
		// Add further code here as needed
	}

	var token, tokenErr = loadToken()

	if tokenErr != nil {
		fmt.Println("Error getting token: %v\n", tokenErr)
	}

	workspaceId, workspaceName, err := getWorkspaces(token)
	repoName, repoId, err := linkRepo(token, workspaceId)

	fmt.Printf("\033[1;32m✓ Syncing local repository \033[1;36m%s\033[1;32m to remote space \033[1;36m%s\033[0m\n", repoName, workspaceName)
	fmt.Println("\033[1;33m⚠ ALERT: Please leave this open while programming\033[0m")
	if err != nil {
		fmt.Printf("Error getting workspaces: %v\n", err)
		return
	}

	sendDataToServer(dirContent, token, workspaceId, repoName, repoId, false)
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

func installationMethodCommand() (string, error) {
	if _, err := exec.Command("brew", "list", "mordecai").Output(); err == nil {

		fmt.Println("brew")
		return "brew", nil
	}
	return "curl", nil
}
