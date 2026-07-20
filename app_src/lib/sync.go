package lib

import (
	"runtime"
	"sync"
)

var wg sync.WaitGroup

func Go(f func()) {
	wg.Go(f)
}

func Goexit() {
	runtime.Goexit()
}