#!/bin/sh

CPUMODEL=`cat /proc/cpuinfo | grep "model name" | uniq`
MAXCPUS=`cat /proc/cpuinfo | grep processor | wc -l`

# header info

for C in icc gcc clang; do
	FILE=sparselu-single-omp-tasks.$C.bench
	# cleanup
	rm $FILE;
	echo "#" $CPUMODEL ", " $MAXCPUS "cores" >> $FILE;
	for n in $(seq "$MAXCPUS");
		do echo -e "OMP_NUM_THREADS=$n" >> $FILE;
		for i in {1..10};
			do OMP_NUM_THREADS=$n ../bots/bin/sparselu.$C.single-omp-tasks | grep Time >> $FILE;
		done;
	done;
done
