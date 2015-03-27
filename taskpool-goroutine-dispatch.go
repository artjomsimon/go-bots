package main

import (
	//"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
)

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

func (pool *TaskPool) removeWorker() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.goroutineSlots++
}

func (pool *TaskPool) AddTask(workFunc func()) {

	pool.mutex.Lock()
	if pool.goroutineSlots > 0 {
		pool.goroutineSlots--
		pool.mutex.Unlock()
		//fmt.Printf("GoroutineSlots: %d\n", pool.goroutineSlots)
		go func() {
			defer pool.removeWorker()
			workFunc()
		}()
	} else {
		pool.mutex.Unlock()
		workFunc()
	}
}

func (pool *TaskPool) Start() {
}

func (pool *TaskPool) Stop() {
}
