// +build integration

package persistent

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
)

var addresses = os.Getenv("ETCD_ADDRESSES")

func Test_Acquire(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(addresses, ",")})
	if err != nil {
		t.Fatal(err)
	}

	var (
		keyer   = staticKeyer(time.Now().Format("2006-01-02T15:04:05.9999999"))
		sem     = NewSemaphore(clientv3.NewKV(cli), 2, WithKeyer(keyer))
		ctxt    = context.Background()
		attempt = func(successful bool) {
			acquired, err := sem.Acquire(ctxt, "/foo")
			assert.Nil(t, err)
			assert.Equal(t, successful, acquired)
		}
		successfulAttempt = func() { attempt(true) }
		failedAttempt     = func() { attempt(false) }
	)

	successfulAttempt()
	successfulAttempt()
	// third attempt should return false as the limit is 2
	failedAttempt()

	// we move forward in time
	WithKeyer(staticKeyer(time.Now().Format("2006-01-02T15:04:05.9999999")))(sem)

	// two more successful attempts
	successfulAttempt()
	successfulAttempt()
	// third attempt should return false as the limit is 2
	failedAttempt()
}
