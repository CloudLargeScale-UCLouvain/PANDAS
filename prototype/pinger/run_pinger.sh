#!/bin/bash

# Check if we got an argument
if [ $# -eq 0 ]; then
    echo "Provide a topology file as an argument"
    exit 1
fi

topo_file=$1
send="sender"
recv="receiver"
sender_binary="${2}${send}"
receiver_binary="${2}${recv}"
node_ip=$3

num_packets=500
packet_size=100

# Run receivers
while IFS= read -r line
do
    nick=$(echo "$line" | cut -d',' -f1)
    port=$(echo "$line" | cut -d',' -f2)
    ip=$(echo "$line" | cut -d',' -f4)
    node_type=$(echo "$line" | rev | cut -d',' -f1 | rev)
   	node_ip=$(ip -o -4 addr show scope global | awk '!/^[0-9]+: lo:/ {print $4}' | cut -d '/' -f 1)
		echo ${node_ip}
		echo ${ip}	
    # Check if the IP matches
    if [ "$ip" = "$node_ip" ]; then
				echo "test"
				if [ "$node_type" = "builder" ]; then
					  $sender_binary $topo_file $number_packets $packet_size >> ${2}${node_ip}_${port}.txt 2>&1 &
				fi
        if [ "$node_type" = "validator" ]; then
            echo "Launching receiver $nick"
            $receiver_binary -ip=$ip -port=$port >> ${2}${node_type}_${node_ip}_${port}.txt 2>&1 & 
            if [ $? -ne 0 ]; then
                echo "Error running $nick"
                exit 1
            fi
        fi
    fi
done < "$topo_file"

sleep 100
# Check if the sender ran successfully
if [ $? -ne 0 ]; then
    echo "Error running sender"
    exit 1
fi

# Kill all the receivers
pkill -f "$receiver_binary"
if [ $? -ne 0 ]; then
    echo "Error killing receivers"
    exit 1
fi

