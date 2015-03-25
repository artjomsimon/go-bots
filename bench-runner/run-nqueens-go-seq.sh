#!/bin/sh

CPUMODEL=`cat /proc/cpuinfo | grep "model name" | uniq`
MAXCPUS=`cat /proc/cpuinfo | grep processor | wc -l`
BENCHMARK=nqueens

for T in seq gccgo-seq; do
	FILE="$BENCHMARK.go-$T.bench"
	
## cleanup
	rm $FILE
	
	echo "#" $CPUMODEL ", " $MAXCPUS "cores" >> $FILE;
	
	for n in 1;
		do echo -e "OMP_NUM_THREADS=$n" >> $FILE;
		for i in {1..10};
			do OMP_NUM_THREADS=$n ../benchmarks/bin/$BENCHMARK-$T -n 12 | grep -i "time" >> $FILE;
		done;
	done;
done;
