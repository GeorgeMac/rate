GO_FLAGS ?= '-mod=vendor'
ETCD_ADDRESSES=http://localhost:2379

.PHONY: install
install: ## Install rate into Go global bin folder
	@go ${GO_FLAGS} install ./...

.PHONY: build
build: make-bin-dir ## Build rate into local bin/ directory
	@go build ${GO_FLAGS} -o bin/rate ./cmd/rate/.
	@echo "Built rate into bin/rate"

.PHONY: test
test: ## Test all the things
	@go test ${GO_FLAGS} ./...

.PHONY: integration-test
integration-test: ## Run integration tests (requires access to etcd)
	@ETCD_ADDRESSES=${ETCD_ADDRESSES} go test ${GO_FLAGS} -tags integration ./...

.PHONY: docker-integration-test
docker-integration-test: ## Run integration tests using docker to bootstrap and run etcd
	@./hack/integration.sh

.PHONY: deps
deps: ## Fetch and vendor dependencies
	@go get ./...
	@go mod vendor

.PHONY: lint
lint: ## Lint project
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint 2>/dev/null
	@golangci-lint run ./...

.PHONY: todos
todos: ## Print out any TODO comments
	@find . -name "*.go" | grep -v "vendor" | xargs grep -n "TODO" || exit 0

.PHONY: ready-to-submit
ready-to-submit: lint ## Prints a message when the project is ready to be submitted
	@find . -name "*.go" | grep -v "vendor" | xargs grep -n "TODO" >/dev/null || echo "Ready to go ✓"

.PHONY: docker
docker: ## Builds rate into a docker container
	@docker build -t rate .

.PHONY: compose-up
compose-up: compose-build ## Brings up a demonstration of the rate limiter in docker (requires docker + compose)
	@docker-compose up -d

compose-build:
	@docker-compose build

make-bin-dir:
	@mkdir -p bin

.PHONY: attack ## Run an attack against the docker compose created stack
attack: install-vegeta
	@./hack/attack.sh

install-vegeta:
	@echo Installing Vegeta using Go Get
	@go get -u github.com/tsenart/vegeta 2>&1 >/dev/null

# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s%-2s\033[0m %s\n", $$1, "›", $$2}'

.DEFAULT_GOAL := help
