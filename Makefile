.PHONY: all build test run clean

BINARY_NAME=grammar

all: build

build:
	go build -o $(BINARY_NAME) .

test:
	go test -timeout=10s ./... $(ARGS)

run: build
	./$(BINARY_NAME) $(ARGS)

clean:
	go clean
	rm -f $(BINARY_NAME)
