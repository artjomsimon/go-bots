package main

import (
	"os"
	"runtime"
	"strconv"
	"sync"
)

// from https://groups.google.com/forum/#!searchin/golang-nuts/goroutine$20pool/golang-nuts/wnlBR25aFtg/34NAUmZtyA8J
/*
   This function starts off n worker goroutines and allows
you to send work to them.
	In order to close down the work pool, just close the chan that is returned.
	In order to ensure all workers have finished, call Wait() on the returned WaitGroup.
*/

/* Hinweis für treerec-Benchmark: der work-channel blockiert, solange sein puffer (1) voll ist.
   der puffer wird aber nicht geleert, weil dei zu bearbeitete funktion nicht returnt, sondern rekursiv sich selbst aufruft und damit einen neuen task versucht in den channel zu stecken, was einen deadlock hervorruft. */

type TaskPool struct {
	wait sync.WaitGroup
	work chan func()
}

func NewTaskPool() *TaskPool {
	pool := &TaskPool{}
	return pool
}

func (pool *TaskPool) AddTask(workFunc func()) {
	pool.work <- workFunc
}

func (pool *TaskPool) Start() {
	bindThreads := os.Getenv("OMP_PROC_BIND")
	if bindThreads == "TRUE" {
		runtime.LockOSThread()
	}

	numThreads, err := strconv.Atoi(os.Getenv("OMP_NUM_THREADS"))

	if err != nil || numThreads < 1 {
		numThreads = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(numThreads)

	pool.wait.Add(numThreads)
	pool.work = make(chan func(), numThreads)

	for n := numThreads; n > 0; n-- {
		go func() {
			for x := range pool.work {
				x()
			}
			pool.wait.Done()
		}()
	}
}

func (pool *TaskPool) Stop() {
	close(pool.work)
	pool.wait.Wait()
}
