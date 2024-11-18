package main

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//                     _             _
// __   _____ _ __ ___(_) ___  _ __ (_)_ __   __ _
// \ \ / / _ \ '__/ __| |/ _ \| '_ \| | '_ \ / _` |
//  \ V /  __/ |  \__ \ | (_) | | | | | | | | (_| |
//   \_/ \___|_|  |___/_|\___/|_| |_|_|_| |_|\__, |
//                                           |___/

func updateVersion() error {
	latestVersion, err := getLatestVersion()
	if err == nil && compareVersions(latestVersion, version) > 0 {
		m := VersionUpdateModel{
			latestVersion:  latestVersion,
			currentVersion: version,
		}

		p := tea.NewProgram(m)
		finalModel, err := p.Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}

		if finalModel.(VersionUpdateModel).choice == "y" {
			// Checks how the CLI tool was initally installed
			methodOfInstallation, err := installationMethodCommand()
			if err != nil {
				fmt.Println("Error checking installation method:", err)
				return err
			}
			var cmd *exec.Cmd
			var updateMessage string

			switch methodOfInstallation {
			case "brew":
				cmd = exec.Command("brew", "upgrade", "mordecai")
				updateMessage = "Successfully updated Mordecai using Brew"
			case "curl":
				cmd = exec.Command("bash", "-c", "curl -sSL https://raw.githubusercontent.com/codeyarduk/mordecai/main/install.sh | bash")
				updateMessage = "Successfully updated Mordecai using Brew"
			default:
				fmt.Println("Unknown installation method. Unable to update.")
				return err

			}

			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error executing command: %v\n", err)
				return err
			}

			fmt.Printf("\n%s\n%s\n", output, updateMessage)

			os.Exit(0)
		}

		return err
	}

	return nil
}

func getLatestVersion() (string, error) {

	// Check for the latest version
	type Release struct {
		TagName string `json:"tag_name"`
	}

	// Make the HTTP GET request
	resp, err := http.Get(githubAPI)
	if err != nil {
		fmt.Printf("Error fetching release: %v\n", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected response status: %d\n", resp.StatusCode)
	}

	// Read and parse the response body
	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
	}

	return release.TagName, err
}

func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		n1, _ := strconv.Atoi(v1Parts[i])
		n2, _ := strconv.Atoi(v2Parts[i])
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return len(v1Parts) - len(v2Parts)
}

type VersionUpdateModel struct {
	latestVersion  string
	currentVersion string
	choice         string
	quitting       bool
}

func (m VersionUpdateModel) Init() tea.Cmd {
	return nil
}

func (m VersionUpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.choice = "y"
			m.quitting = true
			return m, tea.Quit
		case "n", "N", "q", "Q", "ctrl+c":
			m.choice = "n"
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m VersionUpdateModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	s.WriteString(statusMessageStyle.Render("  UPDATE AVAILABLE  "))
	s.WriteString("\n\n")

	s.WriteString("Current version: ")
	s.WriteString(versionStyle.Render(m.currentVersion))
	s.WriteString("\n")

	s.WriteString("Latest version:  ")
	s.WriteString(versionStyle.Render(m.latestVersion))
	s.WriteString("\n\n")

	s.WriteString("To continue using the CLI tool, we need to update it.\n")
	s.WriteString("Can we install the update? (y/N): ")
	s.WriteString(versionStyle.Render(m.choice))

	return s.String()
}

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	divider = lipgloss.NewStyle().
		Foreground(subtle).
		Render("•")

	urlStyle = lipgloss.NewStyle().Foreground(special).Underline(true)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#FF5F87")).
				Padding(0, 1)

	versionStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)
)
