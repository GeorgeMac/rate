package rate

import (
	"context"
	"net/http"
)

// Acquirer is a type which can authorize a key to be actionable
// It should return true if a caller is allowed to proceed
// and action a request given the provided key at the point in time
// The effect of which is to impose limits for given keys when
// work can take place
type Acquirer interface {
	Acquire(key string) (bool, error)
}

// Sleeper sleeps the current routine until
// a configure point in the future or until
// the provided context is cancelled
type Sleeper interface {
	Sleep(context.Context)
}

// Limiter is a http.Handler which limits incoming requests using
// based on the response of a Acquirer per request path
type Limiter struct {
	proxy    http.Handler
	acquirer Acquirer
	sleeper  Sleeper
}

// NewLimiter constructs a newly configured requirer
func NewLimiter(proxy http.Handler, acquirer Acquirer, opts ...Option) Limiter {
	l := Limiter{
		proxy:    proxy,
		acquirer: acquirer,
	}

	Options(opts).Apply(&l)

	return l
}

// ServeHTTP handles the provided request by imposing any active
// limits on the request path and then delegating the request to
// the underlying proxy
func (l Limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for {
		// check if request is ready to be served
		acquired, err := l.acquirer.Acquire(r.URL.Path)
		if err != nil {
			http.Error(w, "service currently unavailable", http.StatusServiceUnavailable)
			return
		}

		if acquired {
			break
		}

		// given the context has not been cancelled
		// e.g. client closed connection
		select {
		case <-r.Context().Done():
			return
		default:
		}

		// sleep using the configured sleep until ready
		l.sleeper.Sleep(r.Context())
	}

	// delegate to proxy handler
	l.proxy.ServeHTTP(w, r)
}
