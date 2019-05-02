package persistent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/oklog/ulid"
	"go.etcd.io/etcd/clientv3"
)

var (
	now     = func() time.Time { return time.Now().UTC() }
	entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

	// ErrKeyNotFound is returned when a key cannot be found within etcd
	ErrKeyNotFound = errors.New("key not found")
)

// Semaphore is a type which is backed by etcd key-value store
// it enforces a certain limit of acquisitions for a provided key
// per a defined interval of time
type Semaphore struct {
	kv clientv3.KV

	limit            int
	intervalDuration time.Duration
}

// NewSemaphore returns a configured etcd backed Semaphore which implements rate.Acquirer
func NewSemaphore(kv clientv3.KV, limit int) *Semaphore {
	return &Semaphore{kv: kv, limit: limit, intervalDuration: 1 * time.Minute}
}

// Acquire attempts to acquire a "token" or "slot" within etcd
// for the provided key
// If successful is returns true and the caller can proceed safely
// If the limit has been reached for this current interval this method
// returns false and the caller should try again later
func (s *Semaphore) Acquire(ctxt context.Context, key string) (bool, error) {
	select {
	case <-ctxt.Done():
		return false, ctxt.Err()
	default:
	}

	var (
		prefix      = claimPrefix(key, s.intervalDuration)
		count, err  = s.getInt64(ctxt, prefix)
		countExists = true
	)

	if err != nil {
		if err != ErrKeyNotFound {
			return false, err
		}

		countExists = false
	}

	if count >= int64(s.limit) {
		return false, nil
	}

	if claimed, err := s.claim(ctxt, prefix, count, countExists); err != nil {
		// something went wrong attempting to communicate with etcd
		return false, err
	} else if !claimed {
		// a claim was not successful
		// liklihood is the claimPrefix count has changed so we
		// attempt again
		return s.Acquire(ctxt, key)
	}

	return true, nil
}

func (s *Semaphore) claim(ctxt context.Context, prefix string, count int64, exists bool) (claimed bool, err error) {
	countChanged := clientv3.Compare(clientv3.Version(prefix), "=", 0)
	if exists {
		countChanged = clientv3.Compare(clientv3.Value(prefix), "=", fmt.Sprintf("%d", count))
	}

	resp, err := s.kv.Txn(ctxt).
		If(countChanged).
		Then(clientv3.OpPut(prefix, fmt.Sprintf("%d", count+1))).
		Commit()
	if err != nil {
		return false, err
	}

	return resp.Succeeded, nil
}

func (s *Semaphore) getInt64(ctxt context.Context, key string) (int64, error) {
	resp, err := s.kv.Get(ctxt, key)
	if err != nil {
		return 0, err
	}

	if len(resp.Kvs) > 0 {
		var (
			item     = resp.Kvs[0]
			val, err = strconv.ParseInt(string(item.Value), 10, 64)
		)

		return val, err
	}

	return 0, ErrKeyNotFound
}

// claimPrefix returns the key used for a claim in etcd in the format "$key/$current_interval/$id"
func claimPrefix(key string, dur time.Duration) string {
	return fmt.Sprintf("%s/%s", key, currentInterval(dur))
}

func currentInterval(dur time.Duration) string {
	return interval(now(), dur)
}

func interval(now time.Time, dur time.Duration) string {
	return now.Truncate(dur).Format("2006-01-02T15:04:05.999999999")
}
