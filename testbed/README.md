# PANDAS
PANDAS implementation in go


## Building
```
go build main.go
```

## Create Topology
We have a simple file creating a topology `create_topo.sh`. It outputs a `./nodes.csv` file that can be used to run the nodes. 

By default, it creates a topology with 1 builder, 2 validators and 4 regular nodes. 

You can run `create_topoo.sh -v 10 -r 20` to change to number of validators to 10 and regular nodes to 20.

## Running
Use `./run_local.sh <topology_file>` to run the nodes locally.

## Running on Grid5k

use `ssh mapigaglio@access.grid5000.fr` to access g5k gateway

on the gateway use `ssh nancy` to jump to the cluster site

use `git clone https://github.com/datahop/PANDAS.git` to clone Pandas git

run the experiment with `python PANDAS/experiment_launch.py`

at the end of the experiment use `zip -r results.zip results` to create an archive of the results

use `scp mapigaglio@access.grid5000.fr:nancy/pandas1000Log.zip .` to get the result archive locally 
