#!/bin/bash

# Set variables
GITHUB_REPO="codeyarduk/mordecai"
BINARY_NAME="mordecai"

# Determine system
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Set install directory based on OS
if [ "$OS" = "darwin" ] || [ "$OS" = "linux" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ "$OS" = "windows" ]; then
    INSTALL_DIR="$HOME/bin"  # User's home directory bin folder
else
    echo "Unsupported operating system"
    exit 1
fi

# Normalize OS names
if [ "$OS" = "darwin" ]; then
    OS="Darwin"
elif [ "$OS" = "linux" ]; then
    OS="Linux"
fi

# For arm64 architecture
if [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

echo "Detected OS: $OS"
echo "Detected Architecture: $ARCH"

# Determine latest release
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$GITHUB_REPO/releases/latest | grep "tag_name" | cut -d '"' -f 4)

# Construct download URL
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_RELEASE/${BINARY_NAME}_${OS}_${ARCH}.tar.gz"

echo "Downloading from: $DOWNLOAD_URL"

# Download and install
echo "Downloading $BINARY_NAME..."
if ! curl -L $DOWNLOAD_URL | tar xz -C /tmp; then
    echo "Download or extraction failed"
    exit 1
fi

# Check if the binary was successfully extracted
if [ ! -f "/tmp/$BINARY_NAME" ]; then
    echo "Binary not found after download"
    exit 1
fi

# Create install directory if it doesn't exist (using sudo for Linux/macOS)
if [ "$OS" != "windows" ]; then
    sudo mkdir -p "$INSTALL_DIR"
else
    mkdir -p "$INSTALL_DIR"
fi

# Move binary to install directory
if [ "$OS" = "windows" ]; then
    mv /tmp/$BINARY_NAME "$INSTALL_DIR"
else
    sudo mv /tmp/$BINARY_NAME $INSTALL_DIR
    sudo chmod +x $INSTALL_DIR/$BINARY_NAME
fi

echo "$BINARY_NAME installed successfully in $INSTALL_DIR"
echo "Make sure $INSTALL_DIR is in your PATH"

# Additional instructions for Windows users
if [ "$OS" = "windows" ]; then
    echo "For Windows users:"
    echo "1. Ensure $INSTALL_DIR is in your PATH."
    echo "2. You may need to restart your terminal or run 'refreshenv' for changes to take effect."
fi
