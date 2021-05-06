# lint is memory and CPU intensive, so we can limit on CI to mitigate OOM
LINT_GOGC?=off
LINT_CONCURRENCY?=8
# Set timeout for linter
LINT_DEADLINE?=1m0s

.PHONY: build
build: linux-build windows-build darwin-build

.PHONY: clean
clean:
	rm -Rf vendor

.PHONY: windows-build
windows-build:
	GOOS=windows go build ./...

.PHONY: linux-build
linux-build:
	GOOS=linux go build ./...

.PHONY: darwin-build
darwin-build:
	GOOS=darwin go build ./...

.PHONY: lint
lint:
	go mod tidy
	GOGC=$(LINT_GOGC) golangci-lint run --fix --verbose --concurrency $(LINT_CONCURRENCY) --deadline $(LINT_DEADLINE)

.PHONY: test
test: lint
	go test -v ./...
