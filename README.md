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

The tool *conf2gv* can be used to generate GraphVis files for given configurations.
