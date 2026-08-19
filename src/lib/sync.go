package lib

import (
	"runtime"
	"sync"
)

type (
	WaitGroup = sync.WaitGroup
	RWMutex   = sync.RWMutex
	Mutex     = sync.Mutex
)

var wg WaitGroup

func Go(f func()) {
	wg.Go(f)
}

func Goexit() {
	runtime.Goexit()
}

type Task = func()

type WorkerPool struct {
	// WaitGroup
	Pool chan Task
}
