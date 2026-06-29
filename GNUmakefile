default: build

build:
	go build -v ./...

install: build
	go install -v ./...

test:
	go test -v -timeout=120s ./...

regenerate:
	./scripts/regenerate.sh

docs: build
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	tfplugindocs generate --provider-name kvindo

.PHONY: build install test regenerate docs
