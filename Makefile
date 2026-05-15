BINARY := bk
GOFLAGS := -trimpath

.PHONY: build test lint clean

build:
	go build $(GOFLAGS) -o $(BINARY) .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
