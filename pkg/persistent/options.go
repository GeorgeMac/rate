package persistent

import "go.etcd.io/etcd/clientv3"

// Option is a functional option for *Semaphore
type Option func(*Semaphore)

// Options is a slice of Option types
type Options []Option

// Apply calls each options from o on Semaphore s in order
func (o Options) Apply(s *Semaphore) {
	for _, opt := range o {
		opt(s)
	}
}

// WithKeyer overrides the keyer on the Semaphore
// with the one provided
func WithKeyer(keyer Keyer) Option {
	return func(s *Semaphore) {
		s.keyer = keyer
	}
}

// WithLease sets an etcd lease on the Semaphore
func WithLease(lease clientv3.Lease) Option {
	return func(s *Semaphore) {
		s.lease = lease
	}
}
