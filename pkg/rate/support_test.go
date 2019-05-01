package rate

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

type localAcquirer struct {
	count int

	counts map[string]int

	mu sync.Mutex
}

func newLocalAcquirer(count int) *localAcquirer {
	return &localAcquirer{count: count, counts: map[string]int{}}
}

func (a *localAcquirer) clear() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.counts = map[string]int{}
}

func (a *localAcquirer) Acquire(key string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if count, ok := a.counts[key]; ok {
		if count >= a.count {
			return false, nil
		}

		a.counts[key] = count + 1

		return true, nil
	}

	a.counts[key] = 1

	return true, nil
}

type waiter struct {
	wakeUp chan struct{}
}

func newWaiter() waiter {
	return waiter{make(chan struct{})}
}

func (s waiter) wake(count int) {
	for i := 0; i < count; i++ {
		s.wakeUp <- struct{}{}
	}
}

func (s waiter) Wait(c context.Context) {
	select {
	case <-s.wakeUp:
	case <-c.Done():
	}
}

func request(t *testing.T, path string) *http.Request {
	t.Helper()

	url, err := url.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	return &http.Request{URL: url}
}
