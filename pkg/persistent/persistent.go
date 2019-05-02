package persistent

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
	"go.etcd.io/etcd/clientv3"
)

var (
	now     = func() time.Time { return time.Now().UTC() }
	entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
)

type Store struct {
	kv clientv3.KV

	limit int
}

func (s *Store) Acquire(key string) (bool, error) {
	count, version, err := s.countInflight(key)
	if err != nil {
		return false, err
	}

	if count >= int64(s.limit) {
		return false, nil
	}

	if claimed, err := s.claim(key, version); err != nil {
		// something went wrong attempting to communicate with etcd
		return false, err
	} else if !claimed {
		// a claim was not successful
		// liklihood is the claimPrefix count has changed so we
		// attempt again
		return s.Acquire(key)
	}

	return true, nil
}

func (s *Store) countInflight(key string) (count, revision int64, err error) {
	err = errors.New("not implemented")
	return
}

func (s *Store) claim(key string, version int64) (claimed bool, err error) {
	err = errors.New("not implemented")
	return
}

func claimKey(key string, dur time.Duration) string {
	id := ulid.MustNew(ulid.Timestamp(now()), entropy).String()

	return fmt.Sprintf("%s/%s", claimPrefix(key, dur), id)
}

// claimPrefix returns the key used for a claim in etcd in the format "$key/$current_interval/$id"
func claimPrefix(key string, dur time.Duration) string {
	return fmt.Sprintf("%s/%s", key, currentInterval(dur))
}

func currentInterval(dur time.Duration) string {
	return interval(now(), dur)
}

func interval(now time.Time, dur time.Duration) string {
	return now.Truncate(dur).Format("15:04:05.999999999")
}
