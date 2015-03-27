# go-bots
Porting the Barcelona OpenMP Tasks Suite to Go

This is an effort to port the benchmarks of the Barcelona OpenMP Tasks Suite (BOTS) to Go.
The benchmarks available in the BOTS suite are commonly used for scalability benchmarks of C/C++ parallelization libraries (Cilk, TBB, OpenMP, ...)
There are numerous academic papers on parallel computing and scheduler performance referencing the BOTS benchmarks.

This endeavour is in its early stage. So far, only the *SparseLU* and the *NQueens* benchmarks are usable.
