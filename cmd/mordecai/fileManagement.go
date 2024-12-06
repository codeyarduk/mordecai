package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
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

// TreeNode represents a file or directory
type TreeNode struct {
	path     string
	name     string
	isDir    bool
	expanded bool
	children []*TreeNode
}

// Model represents the application state
type Model struct {
	root   *TreeNode
	cursor int
	nodes  []*TreeNode // flattened view of visible nodes
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Build the tree from paths
func printFileTree(paths []string, currentDir string) *TreeNode {
	root := &TreeNode{
		path:     currentDir,
		name:     filepath.Base(currentDir),
		isDir:    true,
		expanded: true,
	}

	for _, path := range paths {
		// Make path relative to current directory
		relPath, err := filepath.Rel(currentDir, path)
		if err != nil {
			continue
		}

		parts := strings.Split(filepath.Clean(relPath), string(filepath.Separator))
		current := root

		for i, part := range parts {
			if part == "" || part == "." {
				continue
			}

			found := false
			for _, child := range current.children {
				if child.name == part {
					current = child
					found = true
					break
				}
			}

			if !found {
				isDir := i < len(parts)-1
				node := &TreeNode{
					path:     filepath.Join(current.path, part),
					name:     part,
					isDir:    isDir,
					expanded: false,
				}
				current.children = append(current.children, node)
				current = node
			}
		}
	}
	return root
}

// Flatten visible nodes for display
func (m *Model) flattenTree() {
	m.nodes = make([]*TreeNode, 0)
	var flatten func(*TreeNode, int)
	flatten = func(node *TreeNode, depth int) {
		m.nodes = append(m.nodes, node)
		if node.expanded && node.isDir {
			for _, child := range node.children {
				flatten(child, depth+1)
			}
		}
	}
	flatten(m.root, 0)
}

func (m Model) View() string {
	var s strings.Builder

	for i, node := range m.nodes {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		indent := strings.Repeat("  ", strings.Count(node.path, string(filepath.Separator)))
		icon := "📄"
		if node.isDir {
			if node.expanded {
				icon = "📂"
			} else {
				icon = "📁"
			}
		}

		s.WriteString(fmt.Sprintf("%s %s %s %s\n", cursor, indent, icon, node.name))
	}

	return s.String()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}
		case "enter", "space":
			if m.nodes[m.cursor].isDir {
				m.nodes[m.cursor].expanded = !m.nodes[m.cursor].expanded
				m.flattenTree()
			}
		}
	}
	return m, nil
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
