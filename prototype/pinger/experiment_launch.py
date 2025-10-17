import logging
import enoslib as en
import os
import datetime
import subprocess
import time
import logging
from icecream import ic
import sys
#Upload launch script to site frontend
def execute_ssh_command(launch_script, login, site):
    # SSH command in the format 'ssh <username>@<hostname> "<command>"'
    ssh_command = f'scp {launch_script} {login}@access.grid5000.fr:{site}'

    try:
        # Execute the SSH command
        result = subprocess.run(ssh_command, shell=True, capture_output=True, text=True)
        print("Script send to frontend")
        # Check if the command was successful
        if result.returncode == 0:
            # Print the output
            print(result.stdout)
        else:
            # Print the error message
            print(result.stderr)

    except subprocess.CalledProcessError as e:
        print(f"Error occurred while executing SSH command: {e}")

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
    nb_validator = int(network_size*prop_validator) - 1
    nb_regular = network_size - nb_validator - nb_builder
    index = 1
    partition[0][0] = nb_builder
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

    #========== Parameters ==========
    #Grid5000 parameters
    login = "mapigaglio" #Grid5000 login
    site = "nancy" #Grid5000 Site See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    cluster = "gros" #Gride5000 Cluster name See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    job_name = "PANDAS-udp-test"

    #Node launch script path
    dir_path = os.path.dirname(os.path.realpath(__file__)) #Get current directory path
    launch_script = dir_path +"/" + "run.sh"

    #Experiment parameters

    parcel_size_list = [512]
    network_size_list = [1000]
    nb_run = 1

    k = 0
    nb_expe = len(network_size_list)*len(parcel_size_list)*nb_run
    nb_cluster_machine = 40 #Number of machine booked on the cluster
    prop_validator = 1
    exp_duration = 100  #In seconds
    batch_experiment_name = "PANDAS-Gossip-"
    #Network parameters
    """
    delay = "10%"
    rate = "1gbit"
    loss = "0%"
    symmetric=True
    """
    walltime_in_s = 1200+(exp_duration+30)*nb_expe
    #========== Create and validate Grid5000 and network emulation configurations ==========
    #Log to Grid5000 and check connection

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
        #en.G5kConf.from_settings(job_name=job_name, walltime="01:00:00")
        .add_network_conf(network)
        .add_machine(roles=["experiment"], cluster=cluster, nodes=nb_cluster_machine, primary_network=network) #Add experiment nodes
        .add_machine(roles=["control"], cluster=cluster, nodes=1, primary_network=network) #Add experiment nodes
        .finalize()
    )

    #Validate Grid5000 configuration
    provider = en.G5k(conf)
    roles, networks = provider.init(force_deploy=False)
    roles = en.sync_info(roles, networks)

    #========== Grid5000 network emulation configuration ==========
    #network parameters
    netem = en.Netem()
    (
    netem.add_constraints("delay 200ms", roles["experiment"], symmetric=True)
    )
    """
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
    #execute_ssh_command(launch_script, login, site)
    
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
        """
        for x in roles["experiment"]:
            results = en.run_command("ip -o -4 addr show scope global | awk '!/^[0-9]+: lo:/ {print $4}' | cut -d '/' -f 1", roles=x)
            print(results[0].payload["stdout"])
            ip_list.append(results[0].payload["stdout"])
        """
        for i in range(len(roles["experiment"])):
            server = roles["experiment"][i]
            ip_address_obj = server.filter_addresses(networks=networks["experiment_network"])[0]
            server_private_ip = ip_address_obj.ip.ip
            ip_list.append(server_private_ip)
        
        print("========== Node partition ==========")
        partition = node_partition(nb_cluster_machine, network_size, 1, prop_validator)

        print("========== Create results directories ==========")
        run_name = batch_experiment_name
        experiment_name = "PANDAS-UDP-Test"
        en.run_command(f"mkdir /home/{login}/results/{experiment_name}", roles=roles["experiment"][0])
        en.run_command(f"mkdir /home/{login}/results/{experiment_name}/NodesFiles", roles=roles["experiment"][0])
        en.run_command(f"touch /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/NodesFiles/keys/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/NodesFiles/keysLog/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/results/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/results/nodeLog/", roles=roles["experiment"][0])
        en.run_command(f"mkdir -p /home/{login}/results/{experiment_name}/Log/", roles=roles["experiment"][0])
        en.run_command(f"cp /home/{login}/PANDAS/pinger/sender /home/{login}/results/{experiment_name}/results/sender", roles=roles["experiment"][0])
        en.run_command(f"cp /home/{login}/PANDAS/pinger/receiver /home/{login}/results/{experiment_name}/results/receiver", roles=roles["experiment"][0])

        print("========== Keys list creation ==========")
        keyCreation(ip_list, partition, roles["experiment"], "/home/"+login+"/results/"+experiment_name+"/NodesFiles/nodes.csv","/home/"+login+"/results/"+experiment_name+"/NodesFiles", login)

        #install sar
        """
        print("========== Go + Experiment repo installation ==========")

        for x in roles["experiment"]:
                with en.actions(roles=x, on_error_continue=True, background=True) as p:
                    p.shell(f"/home/{login}/PANDAS/Launching_scripts/server_install.sh")
                    i += 1
        """
        print("========================================")
        print("========== Launch Experiment ===========")
        print("========================================")
        i = 0
        for i in range(len(roles["experiment"])-1,0,-1):
            time.sleep(1)
            with en.actions(roles=roles["experiment"][i], on_error_continue=True, background=True) as p:
                builder, validator, regular = partition[i]
                p.shell(f"/home/mapigaglio/PANDAS/pinger/run_pinger.sh /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv /home/{login}/results/{experiment_name}/results/ {ip_list[i]} >> /home/{login}/results/{experiment_name}/results/nodeLog/run_sh_output_{ip_list[i]}.txt 2>&1")
        k += 1
        print("Experiment:",k,"/",nb_expe)
        with en.actions(roles=roles["experiment"][i], on_error_continue=True, background=True) as p:
            builder, validator, regular = partition[i]
            p.shell(f"/home/{login}/PANDAS-control/control.sh /home/{login}/results/{experiment_name}/NodesFiles/nodes.csv >> /home/{login}/results/{experiment_name}/results/nodeL    og/run_sh_output_control.txt 2>&1")
    #========== Wait job and and release grid5000 ressources ==========
    # netem.destroy()
    #netem.deploy()
    #netem.validate()
    start = datetime.datetime.now() #Timestamp grid5000 job start
    start_formatted = start.strftime("%H:%M:%S")
    elapsed_time = time.time() - start_time
    remaining_time = int(walltime_in_s - elapsed_time)
    for i in track(range(remaining_time), description=f"Waiting for walltime to finish ({remaining_time} secs left)..."):
        time.sleep(1)
if __name__ == "__main__":
    main()
