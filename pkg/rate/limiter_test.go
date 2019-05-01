package rate

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Limiter(t *testing.T) {
	var (
		proxiedCount int64
		proxy        = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			atomic.AddInt64(&proxiedCount, 1)
		})
		sleeper  = newSleeper()
		acquirer = newLocalAcquirer(3)
		limiter  = NewLimiter(proxy, acquirer, WithSleeper(sleeper))
		req      = request(t, "/foo/bar")
	)

	for i := 0; i < 3; i++ {
		// attempt to serve http request
		limiter.ServeHTTP(nil, req)

		assert.Equal(t, proxiedCount, int64(i+1))
	}

	var (
		ctxt = context.Background()
		done = make(chan struct{})
	)

	go func() {
		limiter.ServeHTTP(nil, req.WithContext(ctxt))
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("request should be still inflight and blocking on Acquire(req.Path)")
	default:
	}

	// clear limits
	acquirer.clear()

	// wake up the sleeper
	sleeper.wake(1)

	// wait for request to be done
	<-done

	for i := 0; i < 2; i++ {
		// attempt to serve http request
		limiter.ServeHTTP(nil, req)

		// 4 requests made so far plus 1 for 0 index == 5 offset
		assert.Equal(t, proxiedCount, int64(i+5))
	}
}

func Test_Limiter_Concurrent(t *testing.T) {
	var (
		fooCount, barCount, fooBarCount int64

		proxy = http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/foo":
				atomic.AddInt64(&fooCount, 1)
			case "/bar":
				atomic.AddInt64(&barCount, 1)
			case "/foo/bar":
				atomic.AddInt64(&fooBarCount, 1)
			}
		})

		sleeper  = newSleeper()
		acquirer = newLocalAcquirer(10)
		limiter  = NewLimiter(proxy, acquirer, WithSleeper(sleeper))
		ctxt     = context.Background()
		paths    = []string{
			"/foo",
			"/bar",
			"/foo/bar",
		}
		wg sync.WaitGroup
	)

	for _, path := range paths {
		req := request(t, path).WithContext(ctxt)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(req *http.Request) {
				defer wg.Done()

				limiter.ServeHTTP(nil, req)
			}(req)
		}
	}

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)

		for _, count := range []*int64{&fooCount, &barCount, &fooBarCount} {
			// each time around we should see 10 request for each of the endpoints /foo, /bar and /foo/bar
			require.Equal(t, int64(10), atomic.LoadInt64(count))
			atomic.StoreInt64(count, 0)
		}

		// clear count caches
		acquirer.clear()

		// each time around we have to wake up 30 less sleeping limiter goroutines
		sleeper.wake(300 - ((i + 1) * 30))
	}

	wg.Wait()
}
