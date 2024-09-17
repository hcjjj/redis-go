// 为等待组添加超时功能

package wait

import (
	"sync"
	"time"
)

// Wait is similar with sync.WaitGroup which can wait with timeout
type Wait struct {
	wg sync.WaitGroup
}

// Add adds delta, which may be negative, to the WaitGroup counter.
func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

// Done decrements the WaitGroup counter by one
func (w *Wait) Done() {
	w.wg.Done()
}

// Wait blocks until the WaitGroup counter is zero.
func (w *Wait) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout blocks until the WaitGroup counter is zero or timeout
// returns true if timeout
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan bool, 1)
	go func() {
		defer close(c)
		w.wg.Wait()
		// 当 WaitGroup 的任务全部完成后，c <- true 会将 true 发送到 channel c，表示 WaitGroup 已经完成
		c <- true
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
