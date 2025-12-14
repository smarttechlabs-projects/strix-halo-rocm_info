#!/bin/bash

# Fix Docker iptables issue
echo "🔧 Fixing Docker iptables configuration..."

# Stop Docker service
echo "Stopping Docker service..."
sudo systemctl stop docker

# Clean up Docker iptables rules
echo "Cleaning up iptables rules..."
sudo iptables -t nat -F
sudo iptables -t nat -X
sudo iptables -t filter -F
sudo iptables -t filter -X

# Restart Docker service
echo "Restarting Docker service..."
sudo systemctl start docker

echo "✅ Docker iptables configuration fixed!"
echo "You can now try running the monitoring stack again."