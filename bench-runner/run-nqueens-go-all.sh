#!/bin/sh

CPUMODEL=`cat /proc/cpuinfo | grep "model name" | uniq`
MAXCPUS=`cat /proc/cpuinfo | grep processor | wc -l`
BENCHMARK=nqueens

for T in const-goroutines-taskwait goroutine-dispatch-taskwait notaskpool; do
	FILE="$BENCHMARK.go-$T.bench"
	
	# cleanup
	rm $FILE
	
	echo "#" $CPUMODEL ", " $MAXCPUS "cores" >> $FILE;
	
	for n in $(seq "$MAXCPUS");
		do echo -e "OMP_NUM_THREADS=$n" >> $FILE;
		for i in {1..10};
			do OMP_NUM_THREADS=$n ../benchmarks/bin/$BENCHMARK-$T -n=12 | grep -i "time" >> $FILE;
		done;
	done;
done;
