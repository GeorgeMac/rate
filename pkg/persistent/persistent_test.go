// +build integration

package persistent

import (
	"context"
	"os"
	"strings"
	"testing"

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
		sem  = NewSemaphore(clientv3.NewKV(cli), 2)
		ctxt = context.Background()
	)

	acquired, err := sem.Acquire(ctxt, "/foo")
	assert.Nil(t, err)
	assert.True(t, acquired)

	acquired, err = sem.Acquire(ctxt, "/foo")
	assert.Nil(t, err)
	assert.True(t, acquired)

	acquired, err = sem.Acquire(ctxt, "/foo")
	assert.Nil(t, err)
	assert.False(t, acquired)
}
