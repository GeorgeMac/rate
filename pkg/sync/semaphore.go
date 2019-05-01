package sync

import "sync"

// Semaphore is a concurrency construct used to issue a bound
// number of tokens to callers. Blocking calls to Get until
// a token becomes available.
type Semaphore struct {
	tokens chan struct{}

	mu *sync.RWMutex

	count int
}

// NewSemaphore constructs a newly configured semaphore with
// the provided token count
func NewSemaphore(count int) *Semaphore {
	sem := &Semaphore{
		tokens: make(chan struct{}, count),
		count:  count,
		mu:     &sync.RWMutex{},
	}

	sem.Refill()

	return sem
}

// Acquire returns true if the current semaphore has capacity
// The act of call Acquire removes one token from the bucket
func (s *Semaphore) Acquire() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	select {
	case <-s.tokens:
		return true, nil
	default:
		// if token currently unavailable
		return false, nil
	}
}

// Refill obtains a write lock and refills the token channel
// for more consumers
// Once refilled it broadcasts to any waiting goroutines to
// re-obtain their locks
func (s *Semaphore) Refill() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < s.count; i++ {
		select {
		case s.tokens <- struct{}{}:
			// attempt to fill up to count tokens
		default:
			// otherwise throw extras away
		}
	}
}
