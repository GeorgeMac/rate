package persistent

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	now = func() time.Time { return time.Now().UTC() }

	// errKeyNotFound is returned when a key cannot be found within etcd
	errKeyNotFound = errors.New("key not found")
)

// Keyer generates a key string suitable for current interval in time for a provided key
type Keyer interface {
	Key(string) (key string, expiresIn time.Duration)
}

// KeyerFunc is a func which implements the Keyer interface
type KeyerFunc func(string) (string, time.Duration)

// Key delegates down to underlying KeyerFunc
func (fn KeyerFunc) Key(key string) (string, time.Duration) { return fn(key) }

// IntervalKeyer is a function which returns a Keyer based
// on the provided duration
// Every duration in interval the keyer will return the next
// interval timestamp in the key
func IntervalKeyer(dur time.Duration) Keyer {
	return KeyerFunc(func(key string) (string, time.Duration) {
		var (
			when        = now().Truncate(dur)
			intervalKey = fmt.Sprintf("%s/%s", key, when.Format("2006-01-02T15:04:05.999999999"))
		)

		return intervalKey, time.Until(when)
	})
}

// Semaphore is a type which is backed by etcd key-value store
// it enforces a certain limit of acquisitions for a provided key
// per a defined interval of time
type Semaphore struct {
	kv    clientv3.KV
	lease clientv3.Lease

	limit int
	keyer Keyer
}

// NewSemaphore returns a configured etcd backed Semaphore which implements rate.Acquirer
func NewSemaphore(kv clientv3.KV, limit int, opts ...Option) *Semaphore {
	s := &Semaphore{
		kv:    kv,
		limit: limit,
		keyer: IntervalKeyer(1 * time.Minute),
	}

	Options(opts).Apply(s)

	return s
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
		prefix, expiresIn = s.keyer.Key(key)
		count, err        = s.getInt64(ctxt, prefix)
		countChanged      = clientv3.Compare(clientv3.Value(prefix), "=", fmt.Sprintf("%d", count))
	)

	if err != nil {
		if err != errKeyNotFound {
			return false, err
		}

		// if the key was not found then the count is
		// effectively zero but we must adjust our
		// comparison in the claim transaction slightly
		// to account for it being missing rather than zero
		countChanged = clientv3.Compare(clientv3.Version(prefix), "=", 0)
	}

	if count >= int64(s.limit) {
		return false, nil
	}

	// put a 2 second timeout on the put operation
	tctxt, cancel := context.WithTimeout(ctxt, 2*time.Second)
	defer cancel()

	put, err := s.putWithLease(ctxt, prefix, fmt.Sprintf("%d", count+1), expiresIn)
	if err != nil {
		return false, err
	}

	resp, err := s.kv.Txn(tctxt).
		If(countChanged).
		Then(put).
		Commit()
	if err != nil {
		return false, err
	}

	if !resp.Succeeded {
		// a claim was not successful
		// this is the claimPrefix count has changed so we
		// attempt again until the limit is reached or we
		// are successful
		return s.Acquire(ctxt, key)
	}

	return true, nil
}

func (s *Semaphore) putWithLease(ctxt context.Context, key, val string, ttl time.Duration) (clientv3.Op, error) {
	if s.lease == nil {
		return clientv3.OpPut(key, val), nil
	}

	// ttl in etcd is in seconds and the minimum is 5
	leaseTTL := int64(ttl / time.Second)
	if leaseTTL < 5 {
		leaseTTL = 5
	}

	resp, err := s.lease.Grant(ctxt, leaseTTL)
	if err != nil {
		return clientv3.Op{}, err
	}

	return clientv3.OpPut(key, val, clientv3.WithLease(resp.ID)), nil
}

func (s *Semaphore) getInt64(ctxt context.Context, key string) (int64, error) {
	// put a 1 second timeout on the get operation
	ctxt, cancel := context.WithTimeout(ctxt, 1*time.Second)
	defer cancel()

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

	return 0, errKeyNotFound
}
