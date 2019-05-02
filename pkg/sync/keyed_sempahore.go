package sync

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrorRefillIntervalNotPermitted is returned by a call to NewKeyedSemaphore
// if the refill interval is <= 0
var ErrorRefillIntervalNotPermitted = errors.New("refill interval must be >= 0")

// KeyedSemaphore issues count tokens per provided key
// and periodically refills semaphore with tokens
// after every refillInterval
type KeyedSemaphore struct {
	store *sync.Map

	count int
}

// NewKeyedSemaphore returns a newly configured KeyedSemaphore
// which can be used to borrow tokens for particular keys
// up to a configured limit count, at any one time.
func NewKeyedSemaphore(count int, refillInterval time.Duration) (KeyedSemaphore, error) {
	sem := KeyedSemaphore{store: &sync.Map{}, count: count}

	if refillInterval <= 0 {
		return sem, ErrorRefillIntervalNotPermitted
	}

	go sem.refillLoop(refillInterval)

	return sem, nil
}

// Acquire retrieves a token for a specific key
// true is returned if a slot is acquired otherwise false is returned
func (s KeyedSemaphore) Acquire(_ context.Context, key string) (bool, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = s.store.Load(key); !ok {
		v, _ = s.store.LoadOrStore(key, NewSemaphore(s.count))
	}

	return v.(*Semaphore).Acquire()
}

func (s KeyedSemaphore) refillLoop(refillInterval time.Duration) {
	for {
		// truncate to next interval
		when := time.Now().Add(refillInterval).Truncate(refillInterval)
		// block until we get to when
		<-time.After(time.Until(when))

		s.refillAll()
	}
}

func (s KeyedSemaphore) refillAll() {
	// TODO could fire off each refill in a goroutine
	// with a wait to potentially reduce latency introduced
	// by contention on semaphore locks
	s.store.Range(func(_, v interface{}) bool {
		sem := v.(*Semaphore)

		sem.Refill()

		return true
	})
}
