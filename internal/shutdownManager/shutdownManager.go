package shutdownManager

import (
	"os"
	"sync"
)

var (
	mut sync.Mutex
	wg sync.WaitGroup
	exitCalls = make([]func(wg *sync.WaitGroup), 0)
)

// Register registers a func() to be called right before quitting.
// A registered function obtains a sync.WaitGroup of which the Done() method
// must be called last
func Register(f func(*sync.WaitGroup)) {
	mut.Lock()
	defer mut.Unlock()
	exitCalls = append(exitCalls, f)
}

// Initiate actually initiates the shutdown
func Initiate() {
	for _, f := range exitCalls {
		wg.Add(1)
		go f(&wg)
	}
	wg.Wait()
	os.Exit(0)
}
