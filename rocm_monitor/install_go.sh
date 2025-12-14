#!/bin/bash

# Install Go script for Ubuntu 24.04

set -e

echo "Go Installation Script for ROCm Monitor"
echo "======================================="

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
   echo "Please don't run this script as root. It will ask for sudo when needed."
   exit 1
fi

# Option 1: Install from Ubuntu repositories (easier but might be older version)
install_from_apt() {
    echo "Installing Go from Ubuntu repositories..."
    sudo apt update
    sudo apt install -y golang-go
    echo "Go installed from apt. Version:"
    go version
}

# Option 2: Install latest from official Go website
install_latest() {
    echo "Installing latest Go from official website..."
    
    # Download latest Go
    GO_VERSION="1.21.5"  # Update this to latest version
    wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    
    # Remove any existing Go installation
    sudo rm -rf /usr/local/go
    
    # Extract and install
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    
    # Add to PATH
    echo "Adding Go to PATH..."
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    
    # Source bashrc
    source ~/.bashrc
    
    echo "Go installed. Version:"
    /usr/local/go/bin/go version
    
    echo ""
    echo "Please run 'source ~/.bashrc' or open a new terminal to use Go."
}

# Menu
echo "Choose installation method:"
echo "1) Install from Ubuntu repositories (recommended for beginners)"
echo "2) Install latest version from go.dev"
echo "3) Exit"

read -p "Enter choice [1-3]: " choice

case $choice in
    1)
        install_from_apt
        ;;
    2)
        install_latest
        ;;
    3)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo "Invalid choice. Exiting..."
        exit 1
        ;;
esac

echo ""
echo "Go installation complete!"
echo "You can now build the ROCm Monitor with: make build"