BINARY := taskrunner
CMD    := ./cmd/taskrunner

.PHONY: build test run lint

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

run: build
	./$(BINARY) -file tasks.json -workers 3

lint:
	go vet ./...
	gofmt -d .
