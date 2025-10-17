#!/bin/bash

# Bash Script to launch Gossipsub PANDAS
# ========== Prerequisites Install ==========
# Install experiment on the grid5000 node for better disk usage
log_directory=$1

#!/bin/bash

# Bash Script to launch Gossipsub PANDAS
# ========== Prerequisites Install ==========
echo "========== Prerequisites Install =========="
# Install experiment on the grid5000 node for better disk usage
cd /tmp
sudo-g5k
# Install Go
wget "https://go.dev/dl/go1.21.6.linux-amd64.tar.gz"
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
# Clone experiment code
cp -r /home/mapigaglio/go-libp2p-pubsub-tracer /tmp/
cd /tmp/go-libp2p-pubsub-tracer

traced -d ${log_directory}/tracer