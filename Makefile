# iobuf Makefile

.PHONY: all fmt vet test bench clean

all: fmt vet test

fmt:
	@output=$$(gofmt -l .); if [ -n "$$output" ]; then echo "gofmt: FAIL"; echo "$$output"; exit 1; else echo "gofmt: OK"; fi

vet:
	@go vet ./...

test:
	@go test -timeout 120s ./...

bench:
	@go test -timeout 120s -run=NONE -bench=. -benchmem ./...

clean:
	@rm -f coverage.out
	@go clean -testcache
