import logging
import enoslib as en
import os
import datetime
import subprocess
import time
import logging
from icecream import ic
import sys
from rich.console import Console
from rich.progress import track

console = Console()


#Get timestamp after end of experiment
def add_time(original_time, hours=0, minutes=0, seconds=0):
    time_delta = datetime.timedelta(hours=hours, minutes=minutes, seconds=seconds)
    new_time = original_time + time_delta
    return new_time

#Convert time in second in hh:mm:ss
def seconds_to_hh_mm_ss(seconds):
    hours = seconds // 3600
    minutes = (seconds % 3600) // 60
    seconds = seconds % 60
    return f"{hours:02d}:{minutes:02d}:{seconds:02d}"

#Experiment node partition between Grid5000 machine
def node_partition(nb_cluster_machine, network_size, nb_builder, prop_validator):
    partition = [[0,0,0] for i in range(nb_cluster_machine)]
    nb_validator = int(network_size*prop_validator)
    nb_regular = network_size - nb_validator - nb_builder
    index = 1
    partition[0][0] = 1
    while nb_validator > 0 or nb_regular > 0:
        if index == len(partition):
            index  = 1
        if nb_validator > 0:
            partition[index][1] += 1
            nb_validator -= 1
        elif nb_regular > 0:
            partition[index][2] += 1
            nb_regular -= 1
        index +=1
    return partition

def keyCreation(ip_list, partition, role, nodes_file, key_directory, login):

    index_ip = 0
    port = 10200
    additional_port = 11200
    for i in range(len(role)):
        #results = en.run_command("ip -o -4 addr show scope global | awk '!/^[0-9]+: lo:/ {print $4}' | cut -d '/' -f 1", roles=x)
        #print(results[0].payload["stdout"])
        #for ip in :
            #print(ip)
            #if results[0].payload["stdout"] == ip:

                if partition[i][0] != 0:
                    with en.actions(roles=role[i], on_error_continue=True, background=False) as p:
                        p.shell(f"/home/{login}/PANDAS/create_topo_multi.sh {ip_list[i]} {port} builder {nodes_file} {key_directory}/keys/ {partition[i][0]} {login} {additional_port} >> {key_directory}/keysLog/create_keys_{ip_list[i]}.txt 2>&1 ")
                    port += partition[i][0]
                    additional_port += partition[i][0]
                if partition[i][1] != 0:
                    with en.actions(roles=role[i], on_error_continue=True, background=True) as p:
                        p.shell(f"/home/{login}/PANDAS/create_topo_multi.sh {ip_list[i]} {port} validator {nodes_file} {key_directory}/keys/ {partition[i][1]} {login} {additional_port} >> {key_directory}/keysLog/create_keys_{ip_list[i]}.txt 2>&1 ")
                    port += partition[i][1]
                    additional_port += partition[i][1]
                if partition[i][2] != 0:
                    with en.actions(roles=role[i], on_error_continue=True, background=False) as p:
                        p.shell(f"/home/{login}/PANDAS/create_topo_multi.sh {ip_list[i]} {port} regular {nodes_file} {key_directory}/keys/ {partition[i][2]} {login} {additional_port} >> {key_directory}/keysLog/create_keys_{ip_list[i]}.txt 2>&1 ")
                index_ip += 1
                port = 10200
                additional_port = 11200
def main():

    start_time = time.time()
    #========== Grid5000 Parameters ==========
    login = "mapigaglio" #Grid5000 login
    site = "nancy" #Grid5000 Site See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    cluster = "gros" #Gride5000 Cluster name See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    job_name = "PANDAS"

    #==========Experiment parameters==========

    network_size_list = [300]  #Network size for each experiment, [x, y] run 2 experiment, 1 with x nodes and 1 with y nodes
    nb_expe = len(network_size_list)
    nb_cluster_machine = 60 #Number of machine booked on the cluster
    prop_validator = 0.2 #Proportion of validators
    exp_duration = 800  #Experiment duration (in seconds)
    batch_experiment_name = "PANDAS-Gossip-no-tc-"
    walltime_in_s = 2400+(exp_duration+30)*nb_expe #Job duration

    #========== Grid5000 Configuration ==========
    print("============================================")
    print("========== Grid5000 Configuration ==========")
    print("============================================")

    en.init_logging(level=logging.INFO)
    en.check()
    network = en.G5kNetworkConf(type="prod", roles=["experiment_network"], site=site)
    #network = en.G5kNetworkConf(type="kavlan", roles=["experiment_network"], site=site)

    conf = (
        en.G5kConf.from_settings(job_name=job_name, walltime= seconds_to_hh_mm_ss(walltime_in_s))
        .add_network_conf(network)
        .add_machine(roles=["experiment"], cluster=cluster, nodes=nb_cluster_machine, primary_network=network) #Add experiment nodes
        .finalize()
    )

    #Validate Grid5000 configuration
    provider = en.G5k(conf)
    roles, networks = provider.init(force_deploy=False)
    roles = en.sync_info(roles, networks)

    #========== Grid5000 network emulation configuration ==========
    #network parameters
    """ 
    netem = en.Netem()
    (
    netem.add_constraints("delay 200ms", roles["experiment"], symmetric=True)
    )
    
    netem = en.NetemHTB()
    (
        netem.add_constraints(
            src=roles["experiment"],
            dest=roles["experiment"],
            delay=delay,
            rate=rate,
            loss=loss,
            symmetric=symmetric,
        )
    )
    
    #Deploy network emulation
    netem.deploy()
    netem.validate()
    """

    #========== Deploy Experiment ==========
    #Send launch script to Grid5000 site frontend
    
    print("=======================================")
    print("========== Deploy Experiment ==========")
    print("=======================================")

    k = 0
    for network_size in network_size_list:


        #========== Prepare Experiment ==========
        print("========================================")
        print("========== Prepare Experiment ==========")
        print("========================================")

        print("========== Print servers IP order ==========")
        ip_list = []

        for i in range(len(roles["experiment"])):
            server = roles["experiment"][i]
            ip_address_obj = server.filter_addresses(networks=networks["experiment_network"])[0]
            server_private_ip = ip_address_obj.ip.ip
            ip_list.append(server_private_ip)
        
        print("---------- Node partition ----------")
        partition = node_partition(nb_cluster_machine, network_size, 1, prop_validator)
        print("---------- Create results directories ----------")
        run_name = batch_experiment_name
        experiment_name = run_name+"-b1-v"+str(int(network_size*prop_validator))+"-nv"+str(network_size-int(network_size*prop_validator)-1)+"-prs"+str(512)
        en.run_command(f"mkdir -p /home/{login}/results", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/NodesFiles", roles=roles["experiment"][0])
        en.run_command(f"touch /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/NodesFiles/keys/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/NodesFiles/keysLog/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/results/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/results/nodeLog/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/Log/", roles=roles["experiment"][0])
        en.run_command(f"touch /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv", roles=roles["experiment"][0])


        print("---------- Go + Compilation ----------")

        x = roles["experiment"][0]
        with en.actions(roles=x, on_error_continue=True, background=True) as p:
            p.shell(f"/home/{login}/PANDAS/Launching_scripts/server_install.sh {login}")
            
        print("---------- Keys list creation ----------")
        keyCreation(ip_list, partition, roles["experiment"], "/home/"+login+"/results/"+experiment_name+"/NodesFiles/nodes.csv","/home/"+login+"/results/"+experiment_name+"/NodesFiles", login)

        #install sar
        print("---------- Sar installation ----------")
        for i in range(len(roles["experiment"])):
                with en.actions(roles=roles["experiment"][i], on_error_continue=True, background=True) as p:
                    p.shell(f"apt install sysstat -y >> /home/{login}/results/{experiment_name}/results/nodeLog/run_sh_output_{ip_list[i]}.txt 2>&1")
        """ 
        print("---------- tc configuration ----------")
        ip_str = str([str(ip) for ip in ip_list])
        ip_str = ip_str
        ip_str = ip_str.replace(" ","")
        for i in range(len(roles["experiment"])):
            time.sleep(1)
            with en.actions(roles=roles["experiment"][i], on_error_continue=False, background=False) as p:
                p.shell(f"python3 /home/{login}/PANDAS/Launching_scripts/tc_config.py /home/{login}/PANDAS/Latency/latency.json {ip_str} br0 {i} >> /home/{login}/results/{experiment_name}/Log/tc_config_{ip_list[i]}.txt 2>&1")
        
        time.sleep(10)
        """
        print("========================================")
        print("========== Launch Experiment ===========")
        print("========================================")
        i = 0
        for i in range(len(roles["experiment"])):
            time.sleep(1)
            if i == len(roles["experiment"]) - 1:
                with en.actions(roles=roles["experiment"][i], on_error_continue=True, background=True) as p:
                    p.shell(f"/home/{login}/PANDAS/Launching_scripts/run.sh /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv 4 /home/{login}/results/{experiment_name}/NodesFiles/keys/ /home/{login}/results/{experiment_name}/results/ {ip_list[i]} {exp_duration} {login} {ip_list[0]} /home/{login}/results/{experiment_name} >> /home/{login}/results/{experiment_name}/results/nodeLog/run_sh_output_{ip_list[i]}.txt 2>&1")
            else:
                with en.actions(roles=roles["experiment"][i], on_error_continue=True, background=True) as p:
                    p.shell(f"/home/{login}/PANDAS/Launching_scripts/run.sh /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv 4 /home/{login}/results/{experiment_name}/NodesFiles/keys/ /home/{login}/results/{experiment_name}/results/ {ip_list[i]} {exp_duration} {login} {ip_list[0]} /home/{login}/results/{experiment_name} >> /home/{login}/results/{experiment_name}/results/nodeLog/run_sh_output_{ip_list[i]}.txt 2>&1")
        k += 1
        print("Experiment:",k,"/",nb_expe)
        with en.actions(roles=roles["experiment"][i], on_error_continue=True, background=True) as p:
            p.shell(f"/home/{login}/PANDAS-control/control.sh /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv {login} >> /home/{login}/results/{experiment_name}/results/nodeLog/run_sh_output_control.txt 2>&1")
    #========== Wait job and and release grid5000 ressources ==========
    start = datetime.datetime.now() #Timestamp grid5000 job start
    start_formatted = start.strftime("%H:%M:%S")
    console.print("Start: ", start_formatted, style="bold green")
    console.print("Expected End: ", add_time(start, seconds=walltime_in_s).strftime("%H:%M:%S"), style="bold green")
    elapsed_time = time.time() - start_time
    remaining_time = int(walltime_in_s - elapsed_time)
    for i in track(range(remaining_time), description=f"Waiting for walltime to finish ({remaining_time} secs left)..."):
        time.sleep(1)
if __name__ == "__main__":
    main()
