import json
import subprocess
import os
import sys
import ast
import pandas as pd
import time
def json_to_dataframe(file_path):
    with open(file_path, 'r') as file:
        data = json.load(file)
    df = pd.json_normalize(data, 'latency', ['node'], record_prefix='latency_')
    return df


def main():
    """Main function to execute the script."""
    #if len(sys.argv) != 4:
    #    print("Usage: sudo-g5k python3 set_latency.py <latency_file.json> <ip_list> <interface> <node_number>")
    #    print("Example: sudo-g5k python3 set_latency.py latencies.json \"['192.168.1.100', '192.168.1.101']\" eth0 1")
    #    return
    
    json_file = sys.argv[1]
    #ip_list = ast.literal_eval(sys.argv[2])
    ip_list = sys.argv[2].split("[")[-1]
    ip_list = ip_list.split("]")[0]
    ip_list = ip_list.split(",")
    ip_l = [ip.replace("'", "") for ip in ip_list]
    interface = sys.argv[3]
    node_number = int(sys.argv[4])
    
    nb_node = len(ip_list)
    df = json_to_dataframe(json_file)

    node_df = df[df['node'] == node_number].head(nb_node)
    node_df['node'] = ip_list[node_number]
    node_df['interface'] = interface
    node_df = node_df.reset_index(drop=True)
    node_df['dest'] = "127.0.0.0"
    for x in range(nb_node):
        node_df.loc[x, 'dest'] = ip_list[x]
    
    subprocess.run(["sudo-g5k"], check=True)
    time.sleep(2)
    subprocess.run(["sudo", "tc", "qdisc", "add", "dev", interface, "root", "handle", "1:", "htb", "default", "10"], check=True)

    subprocess.run(["sudo", "tc", "class", "add", "dev", node_df.loc[x, 'interface'], "parent", "1:", "classid", "1:10", "htb", "rate", "800mbit"], check=True)
    handle = 20
    for x in range(node_df.shape[0]):
    
        subprocess.run(["sudo", "tc", "class", "add", "dev", node_df.loc[x, 'interface'], "parent", "1:", "classid", "1:"+str(handle), "htb", "rate", "800mbit"], check=True)
        
        try:
            subprocess.run(["sudo", "tc", "qdisc", "add", "dev", node_df.loc[x, 'interface'], "parent", "1:"+str(handle), "handle", str(handle)+":", "netem", "delay", str(node_df.loc[x, 'latency_0'])+"ms"], check=True)
        except subprocess.CalledProcessError as e:
            print(f"Error: Command '{e.cmd}' returned non-zero exit status {e.returncode}.")
            print(f"Full error message:\n{e.stderr}")
        handle += 10
if __name__ == "__main__":
    main()

