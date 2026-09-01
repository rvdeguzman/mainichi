.PHONY: fmt test build

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.repos/*')

test:
	go test ./...

build:
	go build -o mainichi ./cmd/mainichi
