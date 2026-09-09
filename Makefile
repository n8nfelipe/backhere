BINARY := backhere

.PHONY: build test fmt clean

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -buildvcs=false -trimpath -o bin/$(BINARY) .

test:
	go test ./...

fmt:
	gofmt -w main.go internal

clean:
	go clean
