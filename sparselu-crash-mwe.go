package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
)

var matrixSize, submatrixSize int

/***********************************************************************
 * genmat:
 **********************************************************************/
func genmat(M []*[][]float32) {
	var null_entry bool

	/* generating the structure */
	for ii := 0; ii < matrixSize; ii++ {
		for jj := 0; jj < matrixSize; jj++ {
			/* computing null entries */

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
		}
	}
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
}

func sparselu_par_call(BENCH []*[][]float32) {

	fmt.Printf("Computing SparseLU Factorization (%dx%d matrix with %dx%d blocks) ",
		matrixSize, matrixSize, submatrixSize, submatrixSize)
	// #pragma omp parallel
	// #pragma omp single nowait
	// #pragma omp task untied
	var wg sync.WaitGroup

	for kk := 0; kk < matrixSize; kk++ {

		for ii := kk + 1; ii < matrixSize; ii++ {
			if BENCH[ii*matrixSize+kk] != nil {
				for jj := kk + 1; jj < matrixSize; jj++ {
					if BENCH[kk*matrixSize+jj] != nil {
						//#pragma omp task untied firstprivate(kk, jj, ii) shared(BENCH)
						wg.Add(1)
						kk := kk
						ii := ii
						jj := jj

						go func(wg *sync.WaitGroup) {
							defer (*wg).Done()
							if BENCH[ii*matrixSize+jj] == nil {
								subMatrix := make([][]float32, submatrixSize)
								// go-style initializing 2d matrix in a loop
								for i := range subMatrix {
									subMatrix[i] = make([]float32, submatrixSize)
								}
								BENCH[ii*matrixSize+jj] = &subMatrix
							}
							bmod(*BENCH[ii*matrixSize+kk], *BENCH[kk*matrixSize+jj], *BENCH[ii*matrixSize+jj])
						}(&wg)
					}
				}
				wg.Wait()
			}
		}
	}
}

func main() {

	runtime.GOMAXPROCS(47)

	flag.IntVar(&matrixSize, "n", 50, "Matrix size")
	flag.IntVar(&submatrixSize, "m", 100, "Submatrix size")
	flag.Parse()

	var matrixPar []*[][]float32
	sparselu_init(&matrixPar, "Parallel")
	{
		// we're not measuring initialization time, just like BOTS
		sparselu_par_call(matrixPar)
	}
	fmt.Println("Program ended")

}
