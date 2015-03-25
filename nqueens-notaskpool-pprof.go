/*
ported from nqueens.c
	assuming FORCE_TIED_TASKS = FALSE
	MANUAL_CUTOFF = FALSE
	IF_CUTOFF = FALSE
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
)

var solutions = [...]int{
	1,
	0,
	0,
	2,
	10, /* 5 */
	4,
	40,
	92,
	352,
	724, /* 10 */
	2680,
	14200,
	73712,
	365596,
}
var MAX_SOLUTIONS = cap(solutions) // cap() and len() are guaranteed to fit into an int
var total_count int

/*
 * <a> contains array of <n> queen positions.  Returns 1
 * if none of the queens conflict, and returns 0 otherwise.
 */
// instead of passing a pointer, we use go's "slices", which are
// already pointers to array (regions). []int is a slice of the whole array.
func ok(n int, a []rune) bool {

	// i, j declared implicitly in for loops
	var p, q rune

	for i := 0; i < n; i++ {
		p = rune(a[i])

		for j := i + 1; j < n; j++ {
			q = rune(a[j])
			if q == p || q == p-rune(j-i) || q == p+rune(j-i) {
				return false
			}
		}
	}
	return true
}

func nqueens_ser(n int, j int, a []rune, solutions *int) {

	var res int

	if n == j {
		*solutions = 1
		return
	}

	*solutions = 0

	/* try each possible position for queen <j> */
	for i := 0; i < n; i++ {
		/* allocate a temporary array and copy <a> into it */
		a[j] = rune(i)
		if ok(j+1, a) {
			nqueens_ser(n, j+1, a, &res)
			*solutions += res
		}
	}
}

func nqueens(n int, j int, a []rune, solutions *int, depth int) {
	var csols []int

	if n == j {
		/* good solution, count it */
		*solutions = 1
		return
	}

	*solutions = 0
	csols = make([]int, n) // allocates and zeroes (!)
	//memset(csols, 0, ...)

	var wg sync.WaitGroup
	/* try each possible position for queen <j> */
	for i := 0; i < n; i++ {
		wg.Add(1)
		//#pragma omp task untied
		i := i //http://golang.org/doc/faq#closures_and_goroutines
		go func(wg *sync.WaitGroup) {
			defer (*wg).Done()
			b := make([]rune, len(a))
			copy(b, a)
			b[j] = rune(i)
			if ok(j+1, b) {
				nqueens(n, j+1, b, &csols[i], depth)
			}
		}(&wg)
	}
	//#pragma omp taskwait
	wg.Wait()
	for i := 0; i < n; i++ {
		*solutions += csols[i]
	}
}

func find_queens(size int) {
	total_count = 0

	fmt.Printf("Computing N-Queens algorithm (n=%d)\n", size)

	//#pragma omp parallel
	{
		//#pragma omp single
		{
			a := make([]rune, size)
			nqueens(size, 0, a, &total_count, 0)
		}
	}
	fmt.Println(" completed!")
}
func verify_queens(size int) int {

	if size > MAX_SOLUTIONS {
		fmt.Println("size is ", size, "while MAX_SOLUTIONS is ", MAX_SOLUTIONS)
		return -1
	}
	if total_count == solutions[size-1] {
		fmt.Println("OK: total_count is", total_count, " solutions[size-1] is ", solutions[size-1])
		return 0
	}
	fmt.Println("Verification failed! Total count is", total_count, ", expected: ", solutions[size-1])
	return -2
}

func main() {
	bindThreads := os.Getenv("OMP_PROC_BIND")
	if bindThreads == "TRUE" {
		runtime.LockOSThread()
	}

	numThreads, err := strconv.Atoi(os.Getenv("OMP_NUM_THREADS"))

	if err != nil || numThreads < 1 {
		numThreads = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(numThreads)
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	boardSize := flag.Int("n", 8, "Board size")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	start := Wtime_sec()
	find_queens(*boardSize)
	end := Wtime_sec()
	pprof.StopCPUProfile()
	fmt.Printf("Program time: %.6f s\n", end-start)
	verify_queens(*boardSize)
}
