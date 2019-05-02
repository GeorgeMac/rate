GO_FLAGS ?=

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
compose-up: ## Brings up a demonstration of the rate limiter in docker (requires docker + compose)
	@docker-compose up -d

make-bin-dir:
	@mkdir -p bin

.PHONY: attack
attack: install-vegeta build-attack
	@echo Running attack
	@./hack/attack.sh
	@echo Generating Plot
	@cat result.bin | vegeta plot > plot.html
	@open plot.html

install-vegeta:
	@echo Installing Vegeta using Go Get
	@go get -u github.com/tsenart/vegeta

build-attack:
	@docker build -t rate-attack -f ./hack/Dockerfile.attack ./hack/. 2>&1 >/dev/null

# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s%-2s\033[0m %s\n", $$1, "›", $$2}'

.DEFAULT_GOAL := help
