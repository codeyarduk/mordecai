#!/bin/bash

BINARY_NAME="mordecai"

# Determine system
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Set install directory based on OS
if [ "$OS" = "darwin" ] || [ "$OS" = "linux" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ "$OS" = "windows" ]; then
    INSTALL_DIR="$HOME/bin"
else
    echo "Unsupported operating system"
    exit 1
fi

# Remove the binary
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    if [ "$OS" = "windows" ]; then
        rm "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo rm "$INSTALL_DIR/$BINARY_NAME"
    fi
    echo "$BINARY_NAME has been uninstalled from $INSTALL_DIR"
else
    echo "$BINARY_NAME not found in $INSTALL_DIR"
fi
