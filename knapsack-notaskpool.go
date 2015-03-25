package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

type item struct {
	value  int
	weight int
}

type ByValueWeightRatio []item

func (r ByValueWeightRatio) Len() int      { return len(r) }
func (r ByValueWeightRatio) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ByValueWeightRatio) Less(i, j int) bool {
	c := float64(r[i].value)/float64(r[i].weight) -
		float64(r[j].value)/float64(r[j].weight)
	if c < 0 {
		return true
	}
	return false

}

var best_so_far int

//#pragma omp threadprivate(number_of_tasks)
/*
func compare(a *item, b *item) int {
	c := ((float64)(a.value) / (float64)(a.weight)) - ((float64)(b.value) / (float64)(b.weight))

	if c > 0 {
		return -1
	}
	if c < 0 {
		return 1
	}
	return 0
}
*/
func read_input(filename string, items []item, capacity *int, n *int) {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	/* format of the input: #items capacity\n value1 weight1\n ... */
	fmt.Fscanf(file, "%d %d\n", n, capacity)
	//	fmt.Fscanf(file, "%d", capacity)
	items = items[:*n]

	for i := 0; i < *n; i++ {
		fmt.Fscanf(file, "%d %d\n", &items[i].value, &items[i].weight)
	}

	file.Close()

	/* sort the items on decreasing order of value/weight */
	/* cilk2c is fascist in dealing with pointers, whence the ugly cast */
	//qsort(items, *n, sizeof(struct item), (int (*)(const void *, const void *)) compare)
	sort.Sort(ByValueWeightRatio(items))
}

/*
 * return the optimal solution for n items (first is e) and
 * capacity c. Value so far is v.
 */
func knapsack_par(items []item, c int, n int, v int, sol *int, l int) {
	var with, without, best int
	var ub float64

	/* base case: full knapsack or no items */
	if c < 0 {
		*sol = math.MinInt32
		return
	}

	/* feasible solution, with value v */
	if n == 0 || c == 0 {
		*sol = v
		return
	}

	ub = float64(v) + float64(c*items[0].value)/float64(items[0].weight)

	if ub < float64(best_so_far) {
		/* prune ! */
		*sol = math.MinInt32
		return
	}

	/*
	 * compute the best solution without the current item in the knapsack
	 */

	var wg sync.WaitGroup

	//     #pragma omp task untied firstprivate(items,c,n,v,l) shared(without)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer (*wg).Done()
		knapsack_par(items[1:], c, n-1, v, &without, l+1)
	}(&wg)

	wg.Wait()
	/* compute the best solution with the current item in the knapsack */
	//     #pragma omp task untied firstprivate(items,c,n,v,l) shared(with)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer (*wg).Done()
		knapsack_par(items[1:], c-items[0].weight, n-1, v+items[0].value, &with, l+1)
	}(&wg)

	//#pragma omp taskwait
	wg.Wait()

	if with > without {
		best = with
	} else {
		best = without
	}
	fmt.Println("par:  with: ", with, "without: ", without)

	/*
	 * notice the race condition here. The program is still
	 * correct, in the sense that the best solution so far
	 * is at least best_so_far. Moreover best_so_far gets updated
	 * when returning, so eventually it should get the right
	 * value. The program is highly non-deterministic.
	 */
	if best > best_so_far {
		best_so_far = best
	}

	*sol = best
}
func knapsack_seq(items []item, c int, n int, v int, sol *int) {
	var with, without, best int
	var ub float64

	/* base case: full knapsack or no items */
	if c < 0 {
		*sol = math.MinInt32
		return
	}

	/* feasible solution, with value v */
	if n == 0 || c == 0 {
		*sol = v
		return
	}

	ub = float64(v) + float64(c*items[0].value)/float64(items[0].weight)

	if ub < float64(best_so_far) {
		/* prune ! */
		*sol = math.MinInt32
		return
	}
	/*
	 * compute the best solution without the current item in the knapsack
	 */
	knapsack_seq(items[1:], c, n-1, v, &without)

	/* compute the best solution with the current item in the knapsack */
	knapsack_seq(items[1:], c-items[0].weight, n-1, v+items[0].value, &with)

	if with > without {
		best = with
	} else {
		best = without
	}
	fmt.Println("with: ", with, "without: ", without)

	/*
	 * notice the race condition here. The program is still
	 * correct, in the sense that the best solution so far
	 * is at least best_so_far. Moreover best_so_far gets updated
	 * when returning, so eventually it should get the right
	 * value. The program is highly non-deterministic.
	 */
	if best > best_so_far {
		best_so_far = best
	}

	*sol = best
}
func knapsack_main_par(items []item, c int, n int, sol *int) {
	best_so_far = math.MinInt32

	//     #pragma omp parallel
	{
		//        #pragma omp single
		//       #pragma omp task untied
		{
			knapsack_par(items, c, n, 0, sol, 0)
		}

		//        #pragma omp critical
	}
	fmt.Println("Best value for parallel execution is", *sol)
}

func knapsack_main_seq(items []item, c int, n int, sol *int) {
	best_so_far = math.MinInt32

	knapsack_seq(items, c, n, 0, sol)

	fmt.Println("Best value for sequential execution is", *sol)
}

func knapsack_check(sol_seq int, sol_par int) bool {
	if sol_seq == sol_par {
		return true
	} else {
		return false
	}
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

	var n, capacity, sol_par, sol_seq int
	file := flag.String("f", "", "Input file name")
	flag.Parse()
	items := make([]item, 256)
	read_input(*file, items, &capacity, &n)

	knapsack_main_seq(items, capacity, n, &sol_seq)
	start := Wtime_sec()
	knapsack_main_par(items, capacity, n, &sol_par)
	end := Wtime_sec()
	knapsack_check(sol_seq, sol_par)
	fmt.Printf("Program time: %.6f s\n", end-start)
}
