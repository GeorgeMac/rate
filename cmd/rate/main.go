package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/georgemac/rate/pkg/persistent"
	"github.com/georgemac/rate/pkg/rate"
	"github.com/georgemac/rate/pkg/sync"
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
		waiterOption = rate.WithWaiter(rate.NextIntervalWaiter(time.Minute))
		limiter      = rate.NewLimiter(proxy, acquirer, waiterOption)
	)

	checkError(http.ListenAndServe(":"+*port, limiter))
}
