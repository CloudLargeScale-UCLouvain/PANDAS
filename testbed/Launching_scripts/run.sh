#!/bin/bash

# Bash Script to launch Gossipsub PANDAS
# ========== Prerequisites Install ==========
# Install experiment on the grid5000 node for better disk usage
topo_file=$1
binary_path="./libp2p-das"
num_blocks=$2
key_directory=$3
log_directory=$4
duration=$6
login=$7
builder_ip=$8
log_file=$9
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
cp -r /home/${login}/PANDAS /tmp/
cd /tmp
cd PANDAS

ip=$5
echo ${ip}
systemctl start sysstat
sar -A -o ${log_directory}/sar_logs_${ip} 1 $exp_duration >/dev/null 2>&1 &
sleep 1

#sudo apt install tcpdump
#sudo tcpdump -i any -w /home/${login}/results/PANDAS-Gossip--b1-v200-nv799-prs512/Log/${ip}.pcap &
#tc qdisc del dev <interface_name> root
cd /tmp/PANDAS
#check if we got an argument
if [ $# -eq 0 ]; then
    echo "Provide a topology file as an argument"
    exit 1
fi

# Read topo file line by line
bootstrap_node=`grep ",bootstrap" ${topo_file} | cut -d',' -f4`

while IFS= read -r line
do
    nick=`echo "$line" | cut -d',' -f1`
    port=`echo "$line" | cut -d',' -f2`
    udp_port=`echo "$line" | cut -d',' -f3`
    ip=`echo "$line" | cut -d',' -f4`
    node_type=$(echo "$line" | rev | cut -d',' -f1 | rev)
    node_ip=$(ip -o -4 addr show scope global | awk '!/^[0-9]+: lo:/ {print $4}' | cut -d '/' -f 1)
    if [[ "$ip" == "$node_ip" ]]; then
        $binary_path -UDPport=${udp_port} -duration=${duration} -ip=${ip} -port=${port} -nodeType=${node_type} -debug=true -nick=${nick}-${ip} -key=${key_directory}${ip}-${nick}.key -log=${log_directory} -node=${topo_file} >> ${9}/Log/${ip}-${nick}.txt 2>&1&
        if [ $? -ne 0 ]; then
            echo "Error running $nick"
            exit 1
        fi
    fi

done < $topo_file

sleep 600
FAIL=0
# Wait for the nodes to finish
for job in `jobs -p`
do
    wait $job || $FAIL="Fail"
done

if [ "$FAIL" == "0" ];
then
    echo "All jobes finished successfully" 1>&2
else
    echo "Some jobs failed $FAIL" 1>&2
fi
