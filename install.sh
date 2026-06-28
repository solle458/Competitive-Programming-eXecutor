#!/bin/bash

# If no version is specified as a command line argument, fetch the latest version.
if [ -z "$1" ]; then
    VERSION=$(curl -s https://api.github.com/repos/<GITHUB_USER_NAME>/<GITHUB_REPO_NAME>/releases/latest | grep -o '"tag_name": *"[^"]*"' | sed 's/"tag_name": *"//' | sed 's/"//')
    if [ -z "$VERSION" ]; then
        echo "Failed to fetch the latest version"
        exit 1
    fi
else
    VERSION=$1
fi

OS=$(uname -s)
ARCH=$(uname -m)
URL="https://github.com/<GITHUB_USER_NAME>/<GITHUB_REPO_NAME>/releases/download/${VERSION}/<GITHUB_REPO_NAME>_${OS}_${ARCH}.tar.gz"

echo "Start to install."
echo "VERSION=$VERSION, OS=$OS, ARCH=$ARCH"
echo "URL=$URL"

TMP_DIR=$(mktemp -d)
curl -L $URL -o $TMP_DIR/<GITHUB_REPO_NAME>.tar.gz
tar -xzvf $TMP_DIR/<GITHUB_REPO_NAME>.tar.gz -C $TMP_DIR
sudo mv $TMP_DIR/<GITHUB_REPO_NAME> /usr/local/bin/<GITHUB_REPO_NAME>
sudo chmod +x /usr/local/bin/<GITHUB_REPO_NAME>

rm -rf $TMP_DIR

if [ -f "/usr/local/bin/<GITHUB_REPO_NAME>" ]; then
  echo "[SUCCESS] <GITHUB_REPO_NAME> $VERSION has been installed to /usr/local/bin"
else
  echo "[FAIL] <GITHUB_REPO_NAME> $VERSION is not installed."
fi