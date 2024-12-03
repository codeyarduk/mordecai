package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func serverRequest[T any](endpoint string, body interface{}) (T, error) {
	var result T

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return result, fmt.Errorf("error marshaling request body: %v", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return result, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// First, try to decode the response into a generic error structure
	var errorResp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
		if errorResp.Error == "No Access Token" || errorResp.Error == "Expired Token" {
			fmt.Println("Your session has expired or does not exist, please authenticate to continue.")
			authenticate()
			return result, fmt.Errorf("authentication required: %s", errorResp.Error)
		}
	}

	// Reset the response body for the actual result
	resp.Body.Close()
	resp, err = http.Post(endpoint, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return result, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("error decoding response: %v", err)
	}

	return result, nil
}

func getWorkspaces(token string) (string, string, error) {
	fmt.Println("Fetching available workspaces...")
	endpointURL := fmt.Sprintf("https://api.%s/cli/spaces", siteUrl)

	type Workspace struct {
		WorkspaceID   string `json:"spaceId"`
		WorkspaceName string `json:"spaceName"`
	}

	requestBody := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	workspaces, err := serverRequest[[]Workspace](endpointURL, requestBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch workspaces: %v", err)
	}

	workspaceData := make([]struct {
		WorkspaceID   string `json:"spaceId"`
		WorkspaceName string `json:"spaceName"`
	}, len(workspaces))

	for i, w := range workspaces {
		workspaceData[i] = struct {
			WorkspaceID   string `json:"spaceId"`
			WorkspaceName string `json:"spaceName"`
		}{
			WorkspaceID:   w.WorkspaceID,
			WorkspaceName: w.WorkspaceName,
		}
	}

	// Clear the screen and move cursor to top before showing workspace selection
	fmt.Print("\033[2J")
	fmt.Print("\033[H")

	// Create a new workspace model with the enhanced styling
	m := newWorkspaceModel(workspaceData)

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

	requestBody := struct {
		Token       string `json:"token"`
		WorkspaceId string `json:"spaceId"`
	}{
		Token:       token,
		WorkspaceId: workspaceId,
	}

	type Repository struct {
		RepoID   string `json:"contextId"`
		RepoName string `json:"contextName"`
	}

	repos, err := serverRequest[[]Repository](endpointURL, requestBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch repositories: %v", err)
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

func sendDataToServer(files []FileContent, token string, workspaceId string, repoName string, repoId string, update bool) (string, error) {
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
		WorkspaceId: workspaceId,
		Update:      update,
	}

	// Define the response structure
	type Response struct {
		ContextId string `json:"contextId"`
	}

	// Use the serverRequest wrapper
	response, err := serverRequest[Response](endpointURL, postData)
	if err != nil {
		return "", fmt.Errorf("server request failed: %v", err)
	}

	return response.ContextId, nil
}
