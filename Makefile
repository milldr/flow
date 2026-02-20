BINARY := flow
MODULE := github.com/milldr/flow
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X $(MODULE)/internal/cmd.version=$(VERSION)"

.PHONY: build install test lint clean snapshot demo docs

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/flow

install:
	go install $(LDFLAGS) ./cmd/flow

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/

snapshot:
	goreleaser release --snapshot --clean

demo: build
	bash demo-setup.sh
	vhs demo.tape

docs: build
	bash demo-setup.sh
	vhs docs/commands/init.tape
	vhs docs/commands/state.tape
	vhs docs/commands/render.tape
	vhs docs/commands/exec.tape
	vhs docs/commands/list.tape
	vhs docs/commands/delete.tape
