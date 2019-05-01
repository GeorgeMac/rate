package rate

import (
	"context"
	"testing"
	"time"
)

func Test_NextIntervalWaiter(t *testing.T) {
	var (
		waiter   = NextIntervalWaiter(1 * time.Second)
		ctxt     = context.Background()
		finished = make(chan time.Time)
	)

	go func() {
		waiter.Wait(ctxt)
		finished <- time.Now()
	}()

	// this will likely occur before the wait sends
	// due to it waiting till the next second
	// this may not always be the case though
	now := time.Now()

	now.Before(<-finished)
}
