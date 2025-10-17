import os
import sys
from datetime import datetime, timedelta
import matplotlib.pyplot as plt

def parse_timestamp(timestamp):
    """Parse the timestamp string into a datetime object."""
    return datetime.strptime(timestamp, "%H:%M:%S.%f")

def extract_ip_port(line):
    # Find the substring starting with "multiaddr: /ip4/"
    start_index = line.find("multiaddr: /ip4/")
    if start_index == -1:
        return None
    
    # Extract the IP and port part
    start_index += len("multiaddr: /ip4/")
    end_index = line.find("/p2p/", start_index)
    if end_index == -1:
        return None
    
    ip_port = line[start_index:end_index]
    ip, port_str = ip_port.split('/tcp/')
    port = int(port_str) + 1000

    return f"{ip}:{port}"


def extract_ip_portOld(line):
    # Find the part of the line that starts with "addr:"
    addr_index = line.find("addr:")
    if addr_index == -1:
        return None
    
    # Extract the substring starting from "addr:"
    addr_substring = line[addr_index + len("addr:"):]

    # Find the position of the comma after the IP:Port
    comma_index = addr_substring.find(',')

    # Extract the IP:Port part
    if comma_index != -1:
        ip_port = addr_substring[:comma_index].strip()
    else:
        ip_port = addr_substring.strip()

    return ip_port

def process_logs(log_dir):
    builder_log = None
    validator_logs = []

    # Step 1: Identify builder and validator files
    for filename in os.listdir(log_dir):
        if "builder" in filename:
            builder_log = os.path.join(log_dir, filename)
        elif "validator" in filename:
            validator_logs.append(os.path.join(log_dir, filename))
    
    if not builder_log:
        print("Builder log file not found.")
        return
    
    # Step 2: Extract packet send events from builder log
    send_events = {}
    with open(builder_log, 'r') as file:
        print('Processing builder log: ', builder_log)
        for line in file:
            if "Pushing samples to IP:" in line:
                timestamp_str = line.split()[1]
                addr = line.split("UDP addr: ")[1].strip()
                #print('Builder extracted addr: ', addr)
                timestamp = parse_timestamp(timestamp_str)
                if addr not in send_events:
                    send_events[addr] = []
                send_events[addr].append(timestamp)
    
    # Step 3: Extract receipt events from validator logs
    receipt_events = {}
    for log in validator_logs:
        with open(log, 'r') as file:
            print('Processing validator log: ', log)
            addr = None
            for line in file:
                if "my own ID:" in line: 
                    addr = extract_ip_port(line)
                    #print('validator addr: ', addr)
                if "Got all the seed samples" in line:
                    timestamp_str = line.split()[1]
                    timestamp = parse_timestamp(timestamp_str)
                    if addr not in receipt_events:
                        receipt_events[addr] = []
                    receipt_events[addr].append(timestamp)
    #print("Send events: ", send_events)
    #print("Receipt events: ", receipt_events)
    # Step 4: Match send and receipt events and determine success/failure
    results = []
    total_sends = 0
    total_successes = 0
    total_send_time = 0.0
    send_times_per_node = {}

    for addr, sends in send_events.items():
        total_sends += len(sends)
        receipts = receipt_events.get(addr, [])
        for send in sends:
            matched_receipt = None
            for receipt in receipts:
                if receipt >= send:
                    matched_receipt = receipt
                    receipts.remove(receipt)
                    break
            success = matched_receipt is not None
            results.append((send, addr, "Success" if success else "Failure"))
            if success:
                total_successes += 1
                send_time = (matched_receipt - send).total_seconds()
                total_send_time += (matched_receipt - send).total_seconds()
                if addr not in send_times_per_node:
                    send_times_per_node[addr] = []
                send_times_per_node[addr].append(send_time)

    # Step 5: Output results in chronological order
    results.sort(key=lambda x: x[0])
    for result in results:
        print(f"{result[0]} Pushing samples to UDP addr: {result[1]} - {result[2]}")
    
    # Step 6: Calculate and print average success rate and average send time
    if total_sends > 0:
        average_success_rate = total_successes / total_sends * 100
    else:
        average_success_rate = 0
    
    if total_successes > 0:
        average_send_time = total_send_time / total_successes
    else:
        average_send_time = 0.0
    
    print(f"Average success rate: {average_success_rate:.2f}%")
    print(f"Average send time: {average_send_time} seconds")
    
    # Plotting the distribution of send times per node
    plt.figure(figsize=(10, 6))
    node_addresses = list(send_times_per_node.keys())
    data = list(send_times_per_node.values())
    
    plt.boxplot(data, labels=node_addresses)
    plt.xlabel('Node IP:Port')
    plt.ylabel('Send Time (seconds)')
    plt.title('Distribution of Send Times per Node WITHOUT Pending-Requests')
    plt.xticks(rotation=90, fontsize=8)
    plt.tight_layout()
    plt.savefig('send_times.pdf')

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python analyze_logs.py <log_directory>")
        sys.exit(1)

    log_directory = sys.argv[1]
    process_logs(log_directory)
