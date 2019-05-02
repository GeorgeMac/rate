Rate - A rate limiting service
------------------------------------------

Rate is a simple rate limiting service which can be deployed infront of a downstream service.
It is intended to impose a per minute limit on the number of requests started for each distinct resource.

see [design documents](./docs/DESIGN.md) for more complete design thoughts

### Usage

```shell
rate [flags] <proxied_url>

Usage of rate:
  -port string
    	port on which to service rate limiter (default "4040")
  -rpm int
    	requests per minute (default 100)
```

### Development

#### Dependencies

- Go 1.11+
- cmake

see output of `make` for a list of commands

#### Build

see out of `make`:

```shell
install             › Install rate into Go global bin folder
build               › Build rate into local bin/ directory
test                › Test all the things
deps                › Fetch and vendor dependencies
lint                › Lint project
todos               › Print out any TODO comments
ready-to-submit     › Prints a message when the project is ready to be submitted
```

#### Docker + Playground

```
make compose-up # brings up a mini demo

make attack # uses vegeta to hit the demo with requests over time and plots results
```
