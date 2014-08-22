schedule
========

Cache-Oblivious Scheduling of Shared Workloads


## Pre-Build

First you have to install the Go language (http://golang.org). In Ubuntu this can be done like this:

```apt-get install golang```

In order to compile the code with go you have to set GOPATH enviroment variable e.g. like this:

```export GOPATH=$GOPATH:/path_to_the_repo/schedule ```

## Build

Just run the command:

``` ./build.sh ```

to build the programs.

## Graph Visualization

The tool *conf2gv* can be used to generate GraphVis (http://www.graphviz.org) files for given configurations. With

```./conf2gv --config config/running.xml --style name_id > running.gv```

you can create a gv file, which can then be transformed into your prefered image format using the *dot* commandline tool.

```dot -Tsvg running.gv > running.svg```

## Scheduling

The tool *schedule* can be used to generate schedules of a given configuration. Those schedules will full-fill all precedence constraints and, depending on the algorithm you choose, will be optimized for cache usage.

There are four algorithms:

* baseline: Schedules the jobs in a breadth first like way.
* greedy: Tries to  always optimize the next step according to the total maximum bandwith costs.
* heuristic: Uses several heuristics to find a suitable schedule.
* a_star: Uses the A\* algorithm to find the optimal schedule. Please be aware that this algorithm might consume a lot of RAM and to complete.

To create a schedule for your workload, just run the following command:

```./schedule --config config/running.xml --algo greedy```

If your configuration also contains sizes you can use the *--size* option to utilize them also for the executeion of the algorithm.
