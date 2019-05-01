package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/georgemac/rate/pkg/rate"
	"github.com/georgemac/rate/pkg/sync"
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
		port = flag.String("port", "4040", "port on which to service rate limiter")
		rpm  = flag.Int("rpm", 100, "requests per minute")
	)

	flag.Parse()

	target := flag.Arg(0)
	if target == "" {
		printHelp()
		os.Exit(1)
	}

	url, err := url.Parse(target)
	checkError(err)

	var (
		proxy          = httputil.NewSingleHostReverseProxy(url)
		acquirer, aerr = sync.NewKeyedSemaphore(*rpm, time.Minute)
		waiterOption   = rate.WithWaiter(rate.NextIntervalWaiter(time.Minute))
		limiter        = rate.NewLimiter(proxy, acquirer, waiterOption)
	)

	checkError(aerr)

	http.ListenAndServe(":"+*port, limiter)
}
