package main

import (
	"expvar"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/georgemac/rate/pkg/logging"
	"github.com/georgemac/rate/pkg/metrics"
	"github.com/georgemac/rate/pkg/persistent"
	"github.com/georgemac/rate/pkg/rate"
	"github.com/georgemac/rate/pkg/sync"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

func printHelp() {
	fmt.Println("rate [flags] <proxied_url>")
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var (
		port  = flag.String("port", "4040", "port on which to service rate limiter")
		rpm   = flag.Int("rpm", 100, "requests per minute")
		addrs = flag.String("etcd-addresses", "", "addresses for etcd cluster (if left blank an in-memory semaphore is used instead)")
		level = flag.String("log-level", "debug", "logging level")
	)

	flag.Parse()

	target := flag.Arg(0)
	if target == "" {
		printHelp()
		os.Exit(1)
	}

	var (
		logger        = logrus.New()
		logLevel, err = logrus.ParseLevel(*level)
	)
	checkError(err)

	logger.SetLevel(logLevel)

	url, err := url.Parse(target)
	checkError(err)

	logger.Infof("Proxying requests to %q\n", url)

	var (
		proxy    = httputil.NewSingleHostReverseProxy(url)
		acquirer rate.Acquirer
	)

	acquirer, err = sync.NewKeyedSemaphore(*rpm, time.Minute)
	checkError(err)

	if *addrs != "" {
		// if addresses for etcd are configured then construct
		// a client and replace the acquirer with the persistent
		// etcd back implementation
		cli, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(*addrs, ",")})
		checkError(err)

		acquirer = persistent.NewSemaphore(cli.KV, *rpm)
	}

	var (
		provider     = provider.NewExpvarProvider()
		waiterOption = rate.WithWaiter(rate.NextIntervalWaiter(time.Minute))
		limiter      = rate.NewLimiter(proxy, logging.New(acquirer, logger), waiterOption)
		mux          = http.NewServeMux()
	)

	mux.Handle("/debug/vars", expvar.Handler())
	mux.Handle("/", metrics.Handler(limiter, provider))

	checkError(http.ListenAndServe(":"+*port, mux))
}
