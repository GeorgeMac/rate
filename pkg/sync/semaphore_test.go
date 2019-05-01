package sync

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Semaphore(t *testing.T) {
	var (
		inflight  int64
		semaphore = NewSemaphore(10)
		wg        sync.WaitGroup
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// wait to acquire semaphore token
			acquired, err := semaphore.Acquire()
			require.Nil(t, err)

			for !acquired {
				// sleep between 0 -> 10 millisecond
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

				acquired, err = semaphore.Acquire()
				require.Nil(t, err)
			}

			value := atomic.AddInt64(&inflight, 1)
			if value > 10 {
				// if at any point the inflight count is greater
				// than the semaphore limit then fail the test
				t.Fatal("exeeded semaphore limits")
			}
		}()
	}

	// refill the tokens every 100 milliseconds
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)

		// ensure only 10 processed since last loop
		assert.Equal(t, int64(10), atomic.LoadInt64(&inflight))

		// decrement inflight counter back to zero
		atomic.AddInt64(&inflight, -10)

		// refill tokens
		semaphore.Refill()
	}

	// wait until all semaphore obtaining goroutines exits
	wg.Wait()
}

func Test_KeyedSemaphore_BadInterval(t *testing.T) {
	_, err := NewKeyedSemaphore(10, 0)
	require.Error(t, err, ErrorRefillIntervalNotPermitted)
}
