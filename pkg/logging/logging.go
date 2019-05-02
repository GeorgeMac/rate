package logging

import (
	"context"
	"time"

	"github.com/georgemac/rate/pkg/rate"
	"github.com/sirupsen/logrus"
)

// Acquirer is a logging decorator for other rate.Acquirer types
type Acquirer struct {
	rate.Acquirer

	logger logrus.FieldLogger
}

// New constructs a newly configured Acquirer wrapper
func New(a rate.Acquirer, logger logrus.FieldLogger) Acquirer {
	return Acquirer{a, logger}
}

// Acquire delegates to the embedded Acquirer
// It decorates the call to Acquire with logging before and after it returns
func (a Acquirer) Acquire(ctxt context.Context, key string) (acquired bool, err error) {
	start := time.Now()

	defer func() {
		finish := time.Now()
		a.logger.
			WithField("finish", finish).
			WithField("ellapsed", finish.Sub(start)).
			WithField("acquired", acquired).
			Debugf("Acquire(%q) returned", key)
	}()

	a.logger.WithField("start", start).Debugf("Acquire(%q)", key)

	acquired, err = a.Acquirer.Acquire(ctxt, key)
	return
}
