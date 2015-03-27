package main

import (
	//	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
)

type TaskId chan bool

type TaskPool struct {
	mutex          sync.RWMutex
	goroutineSlots int
}

func NewTaskPool() *TaskPool {
	bindThreads := os.Getenv("OMP_PROC_BIND")
	if bindThreads == "TRUE" {
		runtime.LockOSThread()
	}

	numThreads, err := strconv.Atoi(os.Getenv("OMP_NUM_THREADS"))

	if err != nil || numThreads < 1 {
		numThreads = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(numThreads)

	pool := &TaskPool{
		goroutineSlots: numThreads - 1,
	}
	return pool
}

func (pool *TaskPool) availableGoroutineSlots() int {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	return pool.goroutineSlots
}

func (pool *TaskPool) WaitForTasks(deps []TaskId) {
	for _, dep := range deps {
		if dep != nil {
			pool.WaitForTask(dep)
		}
	}
}

func (pool *TaskPool) WaitForTask(doneChan chan bool) {
	/*	we're stuck in a goroutine, waiting for a child to finish.
		while we wait, free a worker slot. as soon as the child finishes,
		we continue, removing the extra worker slot */
	pool.mutex.Lock()
	pool.goroutineSlots++
	pool.mutex.Unlock()

	/* block while waiting for a done signal */
	<-doneChan

	pool.mutex.Lock()
	pool.goroutineSlots--
	pool.mutex.Unlock()
}

/*	If worker slots are available, run this task as another goroutine.
	If all workers are working, run the task in the current goroutine.
	Return a chan we'll send a done signal to when the task finishes. */
func (pool *TaskPool) AddTask(workFunc func()) chan bool {

	doneChan := make(chan bool, 1)

	pool.mutex.Lock()
	if pool.goroutineSlots > 0 {
		pool.goroutineSlots--
		pool.mutex.Unlock()
		// fmt.Printf("GoroutineSlots: %d\n", pool.goroutineSlots)
		go func() {
			workFunc()
			doneChan <- true
			pool.mutex.Lock()
			pool.goroutineSlots++
			pool.mutex.Unlock()
		}()
	} else {
		pool.mutex.Unlock()
		workFunc()
		doneChan <- true
	}

	return doneChan
}

func (pool *TaskPool) Start() {
}

func (pool *TaskPool) Stop() {
}
