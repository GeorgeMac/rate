package metrics

import (
	"net/http"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
)

var now = time.Now

// Handler decorates a http.Handler and measure latency, inflight gauge
// and a requests handled count
func Handler(handler http.Handler, provider provider.Provider) http.Handler {
	var (
		requestsHandled = provider.NewCounter("requests_handled")
		requestDuration = provider.NewHistogram("request_duration", 5)
		inflight        = provider.NewGauge("requests_inflight")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// begin measuring request duration
		observe := begin(requestDuration)

		inflight.Add(1)

		defer func() {
			// measure latency
			observe()

			// measure request handled
			requestsHandled.Add(1)

			// decrement inflight gauge
			inflight.Add(-1)
		}()

		handler.ServeHTTP(w, r)
	})
}

func begin(histogram metrics.Histogram) func() {
	start := now()
	return func() {
		d := float64(now().Sub(start).Nanoseconds()) / float64(time.Second)
		if d < 0 {
			d = 0
		}

		histogram.Observe(d)
	}
}
