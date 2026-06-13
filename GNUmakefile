default: build

build:
	go build -v ./...

install: build
	go install -v ./...

test:
	go test -v -timeout=120s ./...

regenerate:
	./scripts/regenerate.sh

.PHONY: build install test regenerate
