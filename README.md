# Mordecai CLI

Mordecai is a command-line interface (CLI) tool designed to link your local codebase with a remote workspace and synchronise file changes in real-time.

## Installation

### Brew

```shell
brew tap codeyarduk/mordecai
```

```shell
brew install mordecai
```

### Curl

**Install script**

```shell
curl -sSL https://raw.githubusercontent.com/codeyarduk/mordecai/main/install.sh | bash
```

**Uninstall script**

```shell
curl -sSL https://raw.githubusercontent.com/codeyarduk/mordecai/main/uninstall.sh | bash 
```

## Usage

The basic syntax for using Mordecai is:

```shell 
mordecai <command>
```

### Available Commands

**link** 

```shell 
mordecai link
```

This command:

- Authenticates the user (if not already authenticated)
- Reads the current directory
- Prompts the user to select a remote workspace
- Sends the initial codebase to the selected workspace
- Starts watching the directory for changes and syncs them in real-time

**logout**

```shell
mordecai logout
```

Logs out the current user by deleting the stored authentication token.

**--help**

```shell 
mordecai --help
```

Displays usage information and available commands.

**--version**

```shell 
mordecai --version
```

Displays the version of Mordecai you have got installed

## About the tool

**The basics**

- Authentication: Uses a browser-based OAuth flow for secure user authentication.
- Workspace Selection: Allows users to choose from available remote workspaces.
- File Synchronization: Watches the local directory for changes and syncs them to the remote workspace.
- Gitignore Support: Respects .gitignore rules when scanning directories.
- File Type Filtering: Syncs only specific file types (e.g., .go, .js, .ts, .py, .html, .css, .json, .rb, .md).

**Advanced Concepts**

- Token Management: Securely stores and manages authentication tokens.
- File System Watcher: Utilizes fsnotify for efficient file change detection.
- Debounced Updates: Implements a 5-second debounce to batch file updates.
- Interactive CLI: Uses charmbracelet/bubbles for an enhanced user interface.

**Error Handling**

The CLI provides informative error messages for various scenarios, including authentication failures, file reading errors, and network issues.

**Security Considerations**

- Tokens are stored securely in the user's home directory.
- HTTPS is used for all communications with the server.

**Limitations**

- Currently supports a predefined list of file extensions.
- Requires an active internet connection for synchronization.


For more information or support, please contact [David on Twitter](https://x.com/davidwrossiter).

