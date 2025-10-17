#!/bin/bash

# Bash Script to launch Gossipsub PANDAS
# ========== Prerequisites Install ==========
echo "========== Prerequisites Install =========="

login=$1

# Install experiment on the grid5000 node for better disk usage
cd /tmp
sudo-g5k
# Install Go
wget "https://go.dev/dl/go1.21.6.linux-amd64.tar.gz"
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
# Build experiment code
cd /home/${login}/PANDAS
go build .
sleep 30