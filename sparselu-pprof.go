/*
ported from sparselu_single/sparselu.c
	assuming FORCE_TIED_TASKS = false
	MANUAL_CUTOFF = false
	IF_CUTOFF = false
*/

package main

import (
	"flag"
	"fmt"
	"runtime"
	// DEBUG:
	"log"
	"os"
	"runtime/pprof"
)

var matrixSize, submatrixSize int

/* declaring our own matrix type to avoid triple pointers,
this is necessary because go doesn't offer C-style pointer<->array duality,
so accessing the 2d array passed by a pointer would be unnecessarily troublesome
type Matrix [][]float32
*/

/***********************************************************************
 * checkmat:
 **********************************************************************/
func checkmat(M [][]float32, N [][]float32) bool {
	var r_err, EPSILON float32
	EPSILON = 1e-6

	for i := 0; i < submatrixSize; i++ {
		for j := 0; j < submatrixSize; j++ {
			r_err = M[i][j] - N[i][j]
			if r_err == 0.0 {
				continue
			}

			if r_err < 0.0 {
				r_err = -r_err
			}

			if M[i][j] == 0.0 {
				fmt.Printf("Checking failure: A[%d][%d]=%f  B[%d][%d]=%f; \n",
					i, j, M[i][j], i, j, N[i][j])
				return false
			}
			r_err = r_err / M[i][j]
			if r_err > EPSILON {
				fmt.Printf("Checking failure: A[%d][%d]=%f  B[%d][%d]=%f; Relative Error=%f\n",
					i, j, M[i][j], i, j, N[i][j], r_err)
				return false
			}
		}
	}
	return true
}

/***********************************************************************
 * genmat:
 **********************************************************************/
func genmat(M []*[][]float32) {
	var null_entry bool

	/* generating the structure */
	for ii := 0; ii < matrixSize; ii++ {
		for jj := 0; jj < matrixSize; jj++ {
			/* computing null entries */
			null_entry = false

			if (ii < jj) && (ii%3 != 0) {
				null_entry = true
			}
			if (ii > jj) && (jj%3 != 0) {
				null_entry = true
			}
			if ii%2 == 1 {
				null_entry = true
			}
			if jj%2 == 1 {
				null_entry = true
			}
			if ii == jj {
				null_entry = false
			}
			if ii == jj-1 {
				null_entry = false
			}
			if ii-1 == jj {
				null_entry = false
			}
			/* allocating matrix */
			if null_entry == false {

				// In go, we need to initialize a 2d array by initializing the first dimension and
				// then looping over that, initializing the 2nd dimension: https://golang.org/doc/effective_go.html
				subMatrix := make([][]float32, submatrixSize)
				for i := range subMatrix {
					subMatrix[i] = make([]float32, submatrixSize)
				}

				M[ii*matrixSize+jj] = &subMatrix
				/* error checking not really necessary, because unlike malloc(), make() doesn't simply return "nil" on failure.
				if ((M[ii*matrixSize+jj] == nil)) {
				               bots_message("Error: Out of memory\n");
				               exit(101);
				            }
				*/
				/* initializing matrix */
				init_val := 1325
				for i := 0; i < submatrixSize; i++ {
					for j := 0; j < submatrixSize; j++ {
						init_val = (3125 * init_val) % 65536
						subMatrix[i][j] = (float32)(init_val-32768.0) / 16384.0
						//fmt.Printf("ii=%d\tjj=%d\ti=%d\tj=%d\tsetting content to %.9f\n", ii, jj, i, j, subMatrix[i][j])
					}
				}

			} else {
				M[ii*matrixSize+jj] = nil
			}
		}
	}
}

/***********************************************************************
 * print_structure:
 **********************************************************************/
func print_structure(name string, M []*[][]float32) {
	fmt.Printf("Structure for matrix %s @ %p\n", name, M)
	for ii := 0; ii < matrixSize; ii++ {
		for jj := 0; jj < matrixSize; jj++ {
			if M[ii*matrixSize+jj] != nil {
				fmt.Print("x")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n")
	}
	fmt.Print("\n")
}

/***********************************************************************
 * allocate_clean_block:
 **********************************************************************/
/*
func allocate_clean_block() *float32 {
	var p, q *float32

	p = make(float32, submatrixSize*submatrixSize)
	q = p
	if p != nil {
		for i := 0; i < submatrixSize; i++ {
			for j := 0; j < submatrixSize; j++ {
				*p = 0.0
				p++
			}
		}
	} else {
		fmt.Println("Error: Out of memory")
		exit(101)
	}
	return (q)
}
*/
/***********************************************************************
 * lu0:
 **********************************************************************/
func lu0(diag [][]float32) {

	for k := 0; k < submatrixSize; k++ {
		for i := k + 1; i < submatrixSize; i++ {

			diag[i][k] = diag[i][k] / diag[k][k]
			for j := k + 1; j < submatrixSize; j++ {
				diag[i][j] = diag[i][j] - diag[i][k]*diag[k][j]
			}
		}
	}
}

/***********************************************************************
 * bdiv:
 **********************************************************************/
func bdiv(diag [][]float32, row [][]float32) {
	for i := 0; i < submatrixSize; i++ {
		for k := 0; k < submatrixSize; k++ {
			row[i][k] = row[i][k] / diag[k][k]
			for j := k + 1; j < submatrixSize; j++ {
				row[i][j] = row[i][j] - row[i][k]*diag[k][j]
			}
		}
	}
}

/***********************************************************************
 * bmod:
 **********************************************************************/
func bmod(row [][]float32, col [][]float32, inner [][]float32) {
	for i := 0; i < submatrixSize; i++ {
		for j := 0; j < submatrixSize; j++ {
			for k := 0; k < submatrixSize; k++ {
				inner[i][j] = inner[i][j] - row[i][k]*col[k][j]
			}
		}
	}
}

/***********************************************************************
 * fwd:
 **********************************************************************/
func fwd(diag [][]float32, col [][]float32) {

	for j := 0; j < submatrixSize; j++ {
		for k := 0; k < submatrixSize; k++ {
			for i := k + 1; i < submatrixSize; i++ {
				col[i][j] = col[i][j] - diag[i][k]*col[k][j]
			}
		}
	}
}

func sparselu_init(pBENCH *[]*[][]float32, pass string) {
	*pBENCH = make([]*[][]float32, matrixSize*matrixSize)
	genmat(*pBENCH)
	print_structure(pass, *pBENCH)
}

func sparselu_par_call(BENCH []*[][]float32) {

	fmt.Printf("Computing SparseLU Factorization (%dx%d matrix with %dx%d blocks) ",
		matrixSize, matrixSize, submatrixSize, submatrixSize)
	// #pragma omp parallel
	// #pragma omp single nowait
	// #pragma omp task untied
	for kk := 0; kk < matrixSize; kk++ {
		lu0(*BENCH[kk*matrixSize+kk])
		for jj := kk + 1; jj < matrixSize; jj++ {
			if BENCH[kk*matrixSize+jj] != nil {
				//            #pragma omp task untied firstprivate(kk, jj) shared(BENCH)

				fwd(*BENCH[kk*matrixSize+kk], *BENCH[kk*matrixSize+jj])
			}
		}
		for ii := kk + 1; ii < matrixSize; ii++ {
			if BENCH[ii*matrixSize+kk] != nil {
				//            #pragma omp task untied firstprivate(kk, ii) shared(BENCH)
				bdiv(*BENCH[kk*matrixSize+kk], *BENCH[ii*matrixSize+kk])
			}
		}

		//      #pragma omp taskwait

		for ii := kk + 1; ii < matrixSize; ii++ {
			if BENCH[ii*matrixSize+kk] != nil {
				for jj := kk + 1; jj < matrixSize; jj++ {
					if BENCH[kk*matrixSize+jj] != nil {
						//#pragma omp task untied firstprivate(kk, jj, ii) shared(BENCH)

						if BENCH[ii*matrixSize+jj] == nil {
							subMatrix := make([][]float32, submatrixSize)
							// go-style initializing 2d matrix in a loop
							for i := range subMatrix {
								subMatrix[i] = make([]float32, submatrixSize)
							}
							BENCH[ii*matrixSize+jj] = &subMatrix
						}
						bmod(*BENCH[ii*matrixSize+kk], *BENCH[kk*matrixSize+jj], *BENCH[ii*matrixSize+jj])
					}
				}
				//      #pragma omp taskwait
			}
		}
	}
	fmt.Println(" completed!")
}

func sparselu_seq_call(BENCH []*[][]float32) {
	for kk := 0; kk < matrixSize; kk++ {
		lu0(*BENCH[kk*matrixSize+kk])
		for jj := kk + 1; jj < matrixSize; jj++ {
			if BENCH[kk*matrixSize+jj] != nil {

				fwd(*BENCH[kk*matrixSize+kk], *BENCH[kk*matrixSize+jj])
			}
		}
		for ii := kk + 1; ii < matrixSize; ii++ {
			if BENCH[ii*matrixSize+kk] != nil {

				bdiv(*BENCH[kk*matrixSize+kk], *BENCH[ii*matrixSize+kk])
			}
		}
		for ii := kk + 1; ii < matrixSize; ii++ {
			if BENCH[ii*matrixSize+kk] != nil {
				for jj := kk + 1; jj < matrixSize; jj++ {
					if BENCH[kk*matrixSize+jj] != nil {

						if BENCH[ii*matrixSize+jj] == nil {
							subMatrix := make([][]float32, submatrixSize)
							// go-style initializing 2d matrix in a loop
							for i := range subMatrix {
								subMatrix[i] = make([]float32, submatrixSize)
							}
							BENCH[ii*matrixSize+jj] = &subMatrix
						}
						bmod(*BENCH[ii*matrixSize+kk], *BENCH[kk*matrixSize+jj], *BENCH[ii*matrixSize+jj])
					}
				}
			}
		}
	}
}

func sparselu_fini(BENCH []*[][]float32, pass string) {
	print_structure(pass, BENCH)
}

func sparselu_check(SEQ []*[][]float32, BENCH []*[][]float32) bool {
	var ok = true

	for ii := 0; (ii < matrixSize) && ok; ii++ {
		for jj := 0; (jj < matrixSize) && ok; jj++ {
			if (SEQ[ii*matrixSize+jj] == nil) && (BENCH[ii*matrixSize+jj] != nil) {
				ok = false
			}
			if (SEQ[ii*matrixSize+jj] != nil) && (BENCH[ii*matrixSize+jj] == nil) {
				ok = false
			}
			if (SEQ[ii*matrixSize+jj] != nil) && (BENCH[ii*matrixSize+jj] != nil) {
				ok = checkmat(*SEQ[ii*matrixSize+jj], *BENCH[ii*matrixSize+jj])
			}
		}
	}
	if ok {
		return true
	} else {
		return false
	}
}

func main() {
	//TODO: Move this to schedulers
	runtime.GOMAXPROCS(1)
	matrixSize = *flag.Int("n", 50, "Matrix size")
	submatrixSize = *flag.Int("m", 100, "Submatrix size")
	// DEBUG:
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	// DEBUG:
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	var matrixSeq []*[][]float32
	sparselu_init(&matrixSeq, "Sequential")
	sparselu_seq_call(matrixSeq)

	// DEBUG
	pprof.StopCPUProfile()
	//pool.Start()
	var matrixPar []*[][]float32
	sparselu_init(&matrixPar, "Parallel")
	sparselu_par_call(matrixPar)

	sparselu_fini(matrixPar, "Parallel")
	//pool.Stop()

	sparselu_check(matrixSeq, matrixPar)
}
