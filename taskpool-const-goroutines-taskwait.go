package main

import (
	"container/list"
	//	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
)

type TaskId *list.Element

type TaskPool struct {
	tasks      list.List
	mutex      sync.Mutex
	maxWorkers int
	stopSignal chan bool
}

type taskState byte

const (
	TASK_READY   taskState = iota // http://golang.org/ref/spec#Iota
	TASK_RUNNING taskState = iota
)

type Task struct {
	fn         func()
	state      taskState
	doneSignal chan bool
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

	pool := TaskPool{
		maxWorkers: numThreads - 1,
	}

	pool.tasks.Init()
	runtime.GOMAXPROCS(numThreads)
	pool.stopSignal = make(chan bool, pool.maxWorkers)
	return &pool
}

func (pool *TaskPool) AddTask(fn func()) TaskId {
	task := Task{
		fn:         fn,
		state:      TASK_READY,
		doneSignal: make(chan bool, 1),
	}
	pool.mutex.Lock()
	e := pool.tasks.PushFront(&task)
	pool.mutex.Unlock()
	return e
}

/*

// no function overloading in Go.

func (pool *TaskPool) WaitForTasks(deps ...TaskId) {
	for _, dep := range deps {
		pool.WaitForTask(dep)
	}
}
*/
func (pool *TaskPool) WaitForTasks(deps []TaskId) {
	for _, dep := range deps {
		if dep != nil {
			pool.WaitForTask(dep)
		}
	}
}

func (pool *TaskPool) WaitForTask(dep *list.Element) {

	/*

		check doneChan. if there's no done signal,
		see if anyone already works on my dependency task.
		if not, work on it. if it's already being worked on,
		pick any task in state TASK_READY and execute that one.

	*/

	var depTask *Task
	depTask = dep.Value.(*Task)
	for {
		select {
		case <-depTask.doneSignal:
			return
		default:

			pool.mutex.Lock()
			if depTask.state == TASK_READY {
				depTask.state = TASK_RUNNING
				pool.mutex.Unlock()

				depTask.fn()
				depTask.doneSignal <- true

				pool.mutex.Lock()
				pool.tasks.Remove(dep)
				pool.mutex.Unlock()
				return
			} else {
				pool.mutex.Unlock()
				pool.work(-1)
			}
		}
	}
}

func (pool *TaskPool) Start() {

	for i := pool.maxWorkers; i > 0; i-- {
		//		fmt.Println("starting goroutine", i)
		go func(id int) {
			for {
				pool.work(id)
			}
		}(i)
	}
}

func (pool *TaskPool) work(id int) {
	var t *Task
	pool.mutex.Lock()
	e := pool.tasks.Front()

	if e == nil {
		pool.mutex.Unlock()
		//fmt.Println("yielding on top")
		goto yield
	}

	for {
		//fmt.Print(".")
		t = e.Value.(*Task)
		if t.state != TASK_READY {
			e = e.Next()
			if e == nil {
				pool.mutex.Unlock()
				//fmt.Print("Y")
				goto yield
			}

			t = e.Value.(*Task)
		} else {
			// We found a task ready to be run
			break
		}
	}

	t = e.Value.(*Task)
	// FIXME debug
	//fmt.Print("FOUND task WITH STATUS ", t.state)
	t.state = TASK_RUNNING
	//fmt.Println(" AND SET IT TO ", t.state)
	pool.mutex.Unlock()

	// pick the list element's value and execute it
	//fmt.Printf("Goroutine %d found task, executing\n", id)
	t.fn()
	t.doneSignal <- true
	//close(t.doneSignal)

	// task (and all dependencies) executed, we can safely remove it
	pool.mutex.Lock()
	pool.tasks.Remove(e)
	pool.mutex.Unlock()
yield:
	select {
	case <-pool.stopSignal:
		//		fmt.Println("quitting goroutine ", id)
		return
	default:
		//		fmt.Print("D", id)
		runtime.Gosched()
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
