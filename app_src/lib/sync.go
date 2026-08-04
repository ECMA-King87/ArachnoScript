package lib

import (
	"runtime"
	"sync"
)

type WaitGroup = sync.WaitGroup

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
