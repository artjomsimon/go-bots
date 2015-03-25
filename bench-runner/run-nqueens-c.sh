#!/bin/sh

CPUMODEL=`cat /proc/cpuinfo | grep "model name" | uniq`
MAXCPUS=`cat /proc/cpuinfo | grep processor | wc -l`
FILE=nqueens-omp-tasks.clang.bench


# cleanup
rm $FILE

# header info
echo "#" $CPUMODEL ", " $MAXCPUS "cores" >> $FILE

for n in $(seq "$MAXCPUS");
	do echo -e "OMP_NUM_THREADS=$n" >> $FILE;
	for i in {1..10};
		do OMP_NUM_THREADS=$n ../bots/bin/nqueens.clang.omp-tasks -n 11 | grep Time >> $FILE;
	done;
done
