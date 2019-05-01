package sync

// Token is an empty struct
type Token struct{}

// Semaphore is a concurrency construct used to issue a bound
// number of tokens to callers. Blocking calls to Get until
// a token becomes available.
type Semaphore struct {
	tokens chan Token
}

// New constructs a newly configured semaphore with
// the provided token count
func New(count int) Semaphore {
	tokens := make(chan Token, count)
	for i := 0; i < count; i++ {
		tokens <- Token{}
	}

	return Semaphore{tokens: tokens}
}

// Get blocks until a token is available and then
// returns it to the caller
func (s *Semaphore) Get() Token {
	return <-s.tokens
}

// Release gives the token back to the Semaphore
// to be used again
func (s *Semaphore) Release(t Token) {
	s.tokens <- t
}
