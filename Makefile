# iobuf Makefile

.PHONY: all fmt vet test bench clean

all: fmt vet test

fmt:
	@gofmt -l . | grep -v '^$$' && exit 1 || echo "gofmt: OK"

vet:
	@go vet ./...

test:
	@go test -timeout 30s ./...

bench:
	@go test -run=NONE -bench=. -benchmem ./...

clean:
	@rm -f coverage.out
	@go clean -testcache
