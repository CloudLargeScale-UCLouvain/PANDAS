'''
Code to analyse the success rate of builder seeding operation.
Analyses the builder and validator logs to compute the percentage of successful seeding operations.
A seeding operation involves builder sending packets containing the seed samples.

To run: python3 process_logs.py <path to the Log folder in the G5K results>

'''
import os
import sys
from datetime import datetime, timedelta

def parse_timestamp(timestamp):
    """Parse the timestamp string into a datetime object."""
    return datetime.strptime(timestamp, "%H:%M:%S.%f")

def extract_ip_port(line):
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
                if "myself:" in line: 
                    addr = extract_ip_port(line)
                    #print('validator addr: ', addr)
                if "Got all the seed samples" in line:
                    timestamp_str = line.split()[1]
                    timestamp = parse_timestamp(timestamp_str)
                    if addr not in receipt_events:
                        receipt_events[addr] = []
                    receipt_events[addr].append(timestamp)
    
    # Step 4: Match send and receipt events and determine success/failure
    results = []
    total_sends = 0
    total_successes = 0
    total_send_time = 0.0

    for addr, sends in send_events.items():
        total_sends += len(sends)
        receipts = receipt_events.get(addr, [])
        for send in sends:
            matched_receipt = None
            for receipt in receipts:
                #print("Subtracting ", receipt, " from ", send)
                if abs((send - receipt).total_seconds()) < 10:
                    #print("Success for: ", receipt)
                    matched_receipt = receipt
                    break
            success = matched_receipt is not None
            results.append((send, addr, "Success" if success else "Failure"))
            if success:
                total_successes += 1
                total_send_time += (send-matched_receipt).total_seconds()  # Accumulate the total send time based on the matching receipt

    # Step 5: Output results in chronological order
    results.sort(key=lambda x: x[0])
    for result in results:
        print(f"{result[0]} Pushing samples to UDP addr: {result[1]} - {result[2]}")
    
    # Step 6: Calculate and print average success rate and average send time
    if total_sends > 0:
        average_success_rate = total_successes / total_sends * 100
    else:
        average_success_rate = 0
    
    if total_sends > 0:
        average_send_time = total_send_time / total_sends
    else:
        average_send_time = timedelta()
    
    print(f"Average success rate: {average_success_rate:.2f}%")
    print(f"Average send time: {average_send_time}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python analyze_logs.py <log_directory>")
        sys.exit(1)

    log_directory = sys.argv[1]
    process_logs(log_directory)
