package main

import (
	"flag"
	"fmt"
	"runtime"
)

var pool = NewTaskPool()

var matrixSize, submatrixSize int

/***********************************************************************
 * genmat:
 **********************************************************************/
func genmat(M []*[][]float32) {
	var null_entry bool

	for ii := 0; ii < matrixSize; ii++ {
		for jj := 0; jj < matrixSize; jj++ {
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
			if null_entry == false {

				subMatrix := make([][]float32, submatrixSize)
				for i := range subMatrix {
					subMatrix[i] = make([]float32, submatrixSize)
				}

				M[ii*matrixSize+jj] = &subMatrix

			} else {
				M[ii*matrixSize+jj] = nil
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

func sparselu_init(pBENCH *[]*[][]float32, pass string) {
	*pBENCH = make([]*[][]float32, matrixSize*matrixSize)
	genmat(*pBENCH)
}

func sparselu_par_call(BENCH []*[][]float32) {

	for kk := 0; kk < matrixSize; kk++ {

		for ii := kk + 1; ii < matrixSize; ii++ {
			if BENCH[ii*matrixSize+kk] != nil {
				for jj := kk + 1; jj < matrixSize; jj++ {
					if BENCH[kk*matrixSize+jj] != nil {
						//#pragma omp task untied firstprivate(kk, jj, ii) shared(BENCH)
						jj := jj

						pool.AddTask(func() {
							if BENCH[ii*matrixSize+jj] == nil {
								subMatrix := make([][]float32, submatrixSize)
								// go-style initializing 2d matrix in a loop
								for i := range subMatrix {
									subMatrix[i] = make([]float32, submatrixSize)
								}
								BENCH[ii*matrixSize+jj] = &subMatrix
							}
							bmod(*BENCH[ii*matrixSize+kk], *BENCH[kk*matrixSize+jj], *BENCH[ii*matrixSize+jj])
						})
					}
				}
				pool.Stop()
				pool.Start()
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
		pool.Start()
		sparselu_par_call(matrixPar)
		pool.Stop()
	}
	fmt.Println("Program ended")

}
