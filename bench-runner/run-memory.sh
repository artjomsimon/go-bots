#!/bin/bash

# Print CPU information
CPU="` cat /proc/cpuinfo  | grep -m1 "model name"`"
DATEDIR=`date --iso-8601=seconds`

mkdir $DATEDIR.memory

coproc microbench-c/treerec-parallel 20 100000; 
while test -d /proc/$!; do 
	TIMESTAMP=`date +%s%3N`;
	MEMVAL=`smem -P "^treerec-paralle" | grep -v PID `;
	echo $TIMESTAMP $MEMVAL >> out.memory;
done
