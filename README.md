# Mordecai CLI

Mordecai is a command-line interface (CLI) tool designed to link your local codebase with a remote workspace and synchronise file changes in real-time.

**Installation**

_Coming soon_

**Usage**

The basic syntax for using Mordecai is:

```mordecai <command>```

### Available Commands

**link** 

```mordecai link```

This command:

- Authenticates the user (if not already authenticated)
- Reads the current directory
- Prompts the user to select a remote workspace
- Sends the initial codebase to the selected workspace
- Starts watching the directory for changes and syncs them in real-time

**logout**

```mordecai logout```

Logs out the current user by deleting the stored authentication token.

**--help**

```mordecai --help```

Displays usage information and available commands.

**Key Features**

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


For more information or support, please contact [David on X](https://x.com/davidwrossiter).

