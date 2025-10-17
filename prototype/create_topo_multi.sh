#!/bin/bash
# ========== Parameters ==========
ip=$1
port=$2
node_type=$3
nodes_file=$4
key_directory=$5
nb_node=$6
login=$7
additional_port=$8
cd /tmp
sudo-g5k
# Install Go
wget "https://go.dev/dl/go1.21.6.linux-amd64.tar.gz"
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
cd /home/${login}/PANDAS/keygen
go build .

for i in $(seq 1 $nb_node); do
    nick="${node_type}${port}"
    maddr=`/home/mapigaglio/PANDAS/keygen/keygen ${key_directory}${ip}-${nick}.key ${ip} ${port}`
    echo "${nick},${port},${additional_port},${ip},${maddr},${node_type}" >> ${nodes_file}
    ((port++))
    ((additional_port++))
done