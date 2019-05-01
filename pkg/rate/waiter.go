package rate

import (
	"context"
	"time"
)

// WaiterFunc is a function which implements the
// Waiter interface and delegates to the Wait
// call to the wrapped WaiterFunc
type WaiterFunc func(context.Context)

// Wait delegates to the wrapped WaiterFunc
func (s WaiterFunc) Wait(c context.Context) {
	s(c)
}

// NextIntervalWaiter waits returns a Waiter which blocks
// up until the next configured interval or until the provided
// context.Context is cancelled
func NextIntervalWaiter(interval time.Duration) Waiter {
	return WaiterFunc(func(ctxt context.Context) {
		var (
			now = time.Now()
			// round up to next interval
			when = now.Add(interval).Truncate(interval)
		)

		select {
		case <-time.After(when.Sub(now)):
			// block until next interval
		case <-ctxt.Done():
			// unless context done is closed
		}
	})
}
