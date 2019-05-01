package sync

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Test_Semaphore(t *testing.T) {
	var (
		inflight  int64
		semaphore = New(10)
		wg        sync.WaitGroup
	)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			token := semaphore.Get()

			defer func() {
				// mark that we are no longer inflight
				atomic.AddInt64(&inflight, -1)

				// release token to let other goroutine claim
				semaphore.Release(token)

				// decrement wait group so program will eventually finish
				wg.Done()
			}()

			value := atomic.AddInt64(&inflight, 1)
			if value > 10 {
				t.Fatal("exeeded semaphore limits")
			}

			// sleep between 0 -> 10 millisecond
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		}()
	}

	wg.Wait()
}
