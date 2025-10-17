# 🐼 PANDAS: Simulator and Testnet Experiments

---

## 📘 Table of Contents

1. [Introduction](#introduction)
2. [Hardware](#hardware)
3. [Simulator](#simulator)
   - [Setup and Build](#setup-and-build)
   - [Running Experiments](#running-experiments)
   - [Plotting Results](#plotting-results)
4. [Testnet](#testnet)
   - [Setup and Build](#setup-and-build-1)
   - [Running Experiments](#running-experiments-1)
   - [Plotting Results](#plotting-results-1)
5. [Repository Structure](#repository-structure)

---

## 🧩 Introduction

The artifacts of this work are available at:  
👉 **[CloudLargeScale-UCLouvain/PANDAS](https://github.com/CloudLargeScale-UCLouvain/PANDAS)**

The repository includes:
- The **source code** of PANDAS
- The **scripts** to run the experiments
- The **plotting script** to plot experiments results

Both the simulator and the testnet implementations were run on **Linux-based operating systems**.

---

## Hardware

We run the simulator and a local version of the testbed on a server with 18-core Intel Xeon Gold 5220 CPU and 96 GB
of RAM.

---

## 📁 Repository Structure

```
PANDAS/
├── simulator/     # Source code for large-scale simulator experiments
│   ├── configs/   # Configuration files for network simulations
│   ├── Results/   # Experiment results
│   └── python/    # Plotting scripts
└── testnet/       # Source code for real-world testnet experiments
    ├── results/   # Testnet experiment results
    └── python/    # Log processing and plotting
```

---

## 🧪 Simulator

The **simulator** is implemented in **Java 17** using **Maven** as a build tool.  
It is built on top of [Peersim](http://peersim.sourceforge.net/) and based on the Kademlia implementation by **Daniele Furlan** and **Maurizio Bonani**:  
👉 [http://peersim.sourceforge.net/code/kademlia.zip](http://peersim.sourceforge.net/code/kademlia.zip)

### 🔧 Setup and Build

To install all requirements and build the simulator:

```bash
cd simulator/
sudo chmod +x setup.sh && ./setup.sh
```

This script will:
- Install **Java 17** and **Maven**
- Fetch all dependencies
- Build the simulator

After this, the simulator is ready to run experiments.

---

### ▶️ Running Experiments

The simulator uses configuration files located in `simulator/configs/`.  
Each configuration file corresponds to a specific network size (e.g., `1k.cfg` simulates a network of 1,000 nodes).

To run an experiment:

```bash
cd simulator/
./run.sh <config_file>
```

🗂 Results will be stored in:
- `simulator/Results/` → Experiment results  
- `simulator/logs/` → Detailed logs

---

### 📊 Plotting Results

To plot the experiment results:

```bash
cd simulator/python
python3 plot_results.py <results_folder>
```

The generated plots will be saved in `simulator/python/plots/`

---

## 🌐 Testnet

The **testnet** is implemented in **Go** and built on top of [libp2p](https://libp2p.io/).

For the paper, we run the testbed on [Grid5000](https://www.grid5000.fr/), a **French distributed testbed** for experiment-driven research in computer science.  
It can also be executed locally at a smaller scale.

---

### 🔧 Setup and Build

To install requirements and build the testnet:

```bash
cd testnet/
sudo chmod +x setup.sh && ./setup.sh
```

This script installs:
- The latest version of **Go**
- All required dependencies
- The testnet binaries

After setup, the testnet is ready to run experiments.

---

### ▶️ Running Experiments

#### 1. Create a Topology File

First, generate a `nodes.csv` file defining the network topology (list of nodes and their unique peer IDs):

```bash
./create_topo.sh <number_of_nodes>
```

#### 2. Run the Testnet

Run the network using the topology file:

```bash
./run_local.sh <topology_file>
```

🗂 Results will be stored in:
- `testnet/results/` → Experiment results  
- `testnet/logs/` → Detailed logs

---

### 📊 Plotting Results

To process logs and plot results:

```bash
cd testnet/python
python3 process_logs.py <results_folder>
```

This generates visual analyses and metrics from the testnet experiments.

---
