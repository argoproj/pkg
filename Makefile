# lint is memory and CPU intensive, so we can limit on CI to mitigate OOM
LINT_GOGC?=off
LINT_CONCURRENCY?=8
# Set timeout for linter
LINT_DEADLINE?=1m0s

.PHONY: test
test: lint
	go test -v -covermode=count -coverprofile=coverage.out ./...

.PHONY: lint
lint:
	# golangci-lint does not do a good job of formatting imports
	goimports -local github.com/argoproj/pkg -w `find . ! -path './vendor/*' -type f -name '*.go'`
	GOGC=$(LINT_GOGC) golangci-lint run --fix --verbose --concurrency $(LINT_CONCURRENCY) --deadline $(LINT_DEADLINE)
