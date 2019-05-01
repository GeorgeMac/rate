package rate

// Option is a functional option for the limiter type
type Option func(*Limiter)

// Options is a slice of Option types
type Options []Option

// Apply calls each Option in o in order on the
// provided limiter
func (o Options) Apply(l *Limiter) {
	for _, opt := range o {
		opt(l)
	}
}

// WithSleeper sets the provided sleeper on the limiter
// provided when the returned option is called
func WithSleeper(sleeper Sleeper) Option {
	return func(l *Limiter) {
		l.sleeper = sleeper
	}
}
