#!/usr/bin/env bash

set -e

# Variables
GO_INSTALL_DIR="/usr/local"
PROFILE_FILE="$HOME/.bashrc"

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64) GO_ARCH="amd64" ;;
    aarch64) GO_ARCH="arm64" ;;
    armv6l) GO_ARCH="armv6l" ;;
    armv7l) GO_ARCH="armv6l" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Fetch the latest Go version number
LATEST_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1)
GO_TARBALL="${LATEST_VERSION}.linux-${GO_ARCH}.tar.gz"
GO_URL="https://go.dev/dl/${GO_TARBALL}"

echo "🌀 Installing Go ${LATEST_VERSION} for ${GO_ARCH}..."

# Remove any previous Go installation
sudo rm -rf ${GO_INSTALL_DIR}/go

# Download and extract
curl -LO ${GO_URL}
sudo tar -C ${GO_INSTALL_DIR} -xzf ${GO_TARBALL}
rm ${GO_TARBALL}

# Add Go to PATH if not already present
if ! grep -q 'export PATH=$PATH:/usr/local/go/bin' ${PROFILE_FILE}; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ${PROFILE_FILE}
    echo 'export GOPATH=$HOME/go' >> ${PROFILE_FILE}
    echo 'export PATH=$PATH:$GOPATH/bin' >> ${PROFILE_FILE}
fi

# Reload profile
source ${PROFILE_FILE}

# Verify installation
echo "✅ Go installed successfully!"
go version
