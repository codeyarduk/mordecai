package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

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
	authenticateUrl := fmt.Sprintf("https://%s", siteUrl)
	token, err := loadToken()

	if err != nil {
		return "", fmt.Errorf("failed to load token: %w", err)
	}

	if len(token) > 0 {
		// This is where you will ping the workspaces to see if the token is valid
	}

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", fmt.Errorf("failed to find an available port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	authURL := fmt.Sprintf("%s/cli?port=%d", authenticateUrl, port)

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

type model struct {
	url    string
	choice string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.choice = "y"
			return m, tea.Quit
		case "n", "N":
			m.choice = "n"
			return m, tea.Quit
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Open url: **%s** to authenticate? (y/n): ", m.url)
}

func openBrowser(url string) error {

	p := tea.NewProgram(model{url: url})
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("Bubbletea error: %w", err)
	}

	finalModel := m.(model)
	if finalModel.choice == "y" {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default: // Linux and others
			cmd = exec.Command("xdg-open", url)
		}
		return cmd.Start()
	}
	return fmt.Errorf("user declined to open browser")
}

func startLocalServer(callbackPort int) (string, error) {
	tokenChan := make(chan string, 1)
	errChan := make(chan error, 1)
	redirectUrl := fmt.Sprintf("https://%s/chat", siteUrl)
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		parsedURL, err := url.Parse(r.URL.String())
		if err != nil {
			errChan <- fmt.Errorf("failed to parse URL: %w", err)
			return
		}

		token := parsedURL.Query().Get("token")
		if token != "" {
			saveToken(token)
			tokenChan <- token

			// Immediate redirect
			http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
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
