package main

import (
	"container/list"
	//	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
)

type TaskPool struct {
	tasks         list.List
	mutex         sync.Mutex
	workersActive int
	maxWorkers    int
	stopSignal    chan bool
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

	//	fmt.Println("setting  to ", numThreads)

	pool := &TaskPool{
		maxWorkers: numThreads - 1,
	}

	pool.tasks.Init()
	runtime.GOMAXPROCS(numThreads)
	pool.stopSignal = make(chan bool, pool.maxWorkers)
	return pool
}

func (pool *TaskPool) AddTask(task func()) {
	pool.mutex.Lock()
	if pool.workersActive < pool.maxWorkers {
		pool.tasks.PushFront(task)
		pool.mutex.Unlock()
	} else {
		pool.mutex.Unlock()
		task()
	}
}

func (pool *TaskPool) Start() {
	for i := pool.maxWorkers; i > 0; i-- {
		//fmt.Println("starting goroutine", i)
		go func(id int) {
			for {
				pool.mutex.Lock()
				if pool.tasks.Len() > 0 {
					e := pool.tasks.Front()
					pool.workersActive++
					pool.tasks.Remove(e)
					pool.mutex.Unlock()

					// pick the list element's value and execute it
					//fmt.Printf("Goroutine %d found task, executing\n", id)
					e.Value.(func())()

					pool.mutex.Lock()
					pool.workersActive--
					pool.mutex.Unlock()
				} else {
					pool.mutex.Unlock()
					select {
					case <-pool.stopSignal:
						//fmt.Println("quitting goroutine ", i)
						return
					default:
						//fmt.Println("doing nothing ", i)
						runtime.Gosched()
					}
				}
			}
		}(i)
	}
}

func (pool *TaskPool) Stop() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Send a stop signal to each goroutine
	for n := 0; n < pool.maxWorkers; n++ {
		pool.stopSignal <- true
	}
}
