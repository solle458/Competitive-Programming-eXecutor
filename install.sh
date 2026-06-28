#!/bin/bash

# If no version is specified as a command line argument, fetch the latest version.
if [ -z "$1" ]; then
    VERSION=$(curl -s https://api.github.com/repos/solle458/Competitive-Programming-eXecutor/releases/latest | grep -o '"tag_name": *"[^"]*"' | sed 's/"tag_name": *"//' | sed 's/"//')
    if [ -z "$VERSION" ]; then
        echo "Failed to fetch the latest version"
        exit 1
    fi
else
    VERSION=$1
fi

OS=$(uname -s)
ARCH=$(uname -m)
URL="https://github.com/solle458/Competitive-Programming-eXecutor/releases/download/${VERSION}/cpx_${OS}_${ARCH}.tar.gz"

echo "Start to install."
echo "VERSION=$VERSION, OS=$OS, ARCH=$ARCH"
echo "URL=$URL"

TMP_DIR=$(mktemp -d)
curl -L $URL -o $TMP_DIR/cpx.tar.gz
tar -xzvf $TMP_DIR/cpx.tar.gz -C $TMP_DIR
sudo mv $TMP_DIR/cpx /usr/local/bin/cpx
sudo chmod +x /usr/local/bin/cpx

rm -rf $TMP_DIR

if [ -f "/usr/local/bin/cpx" ]; then
  echo "[SUCCESS] cpx $VERSION has been installed to /usr/local/bin"
else
  echo "[FAIL] cpx $VERSION is not installed."
fi