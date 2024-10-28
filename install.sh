#!/bin/bash

# Set variables
GITHUB_REPO="codeyarduk/mordecai"
BINARY_NAME="mordecai"
INSTALL_DIR="/usr/local/bin"

# Determine system architecture
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Determine latest release
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$GITHUB_REPO/releases/latest | grep "tag_name" | cut -d '"' -f 4)

# Construct download URL
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_RELEASE/${BINARY_NAME}_${OS}_${ARCH}.tar.gz"

# Download and install
echo "Downloading $BINARY_NAME..."
curl -L $DOWNLOAD_URL | tar xz -C /tmp
sudo mv /tmp/$BINARY_NAME $INSTALL_DIR
sudo chmod +x $INSTALL_DIR/$BINARY_NAME

echo "$BINARY_NAME installed successfully!"
