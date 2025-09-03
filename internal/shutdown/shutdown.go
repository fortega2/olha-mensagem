package shutdown

import "sync"

var (
	shutdownChan = make(chan struct{})
	once         sync.Once
)

func Signal() {
	once.Do(func() {
		close(shutdownChan)
	})
}

func Wait() <-chan struct{} {
	return shutdownChan
}
