VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -ldflags "-X github.com/duck-labs/upduck/cmd.version=$(VERSION) -X github.com/duck-labs/upduck/cmd.commit=$(COMMIT) -X github.com/duck-labs/upduck/cmd.date=$(DATE)"

.PHONY: build test clean install help demo run-tower run-server

build:
	go build $(LDFLAGS) -o upduck .

build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/upduck-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/upduck-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/upduck-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/upduck-darwin-arm64 .

lint:
	@echo "==> Running go vet"
	@go vet ./...
	@echo "==> Running staticcheck"
	@staticcheck ./...
	@echo "==> Running gci"
	@gci write -s standard -s default -s localmodule --skip-generated .

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f upduck
	rm -rf dist/
	rm -f coverage.out coverage.html

deps:
	go mod tidy

demo:
	@echo "Creating demo configuration..."
	@mkdir -p ./demo-config
	@echo '{"private_key":"demo-private-key-12345678901234567890123456789012","public_key":"demo-public-key-12345678901234567890123456789012"}' > ./demo-config/wireguard-config.json
	@echo '{"connections":[],"allowed_keys":["demo-server-public-key-123456789012345678901234567890"]}' > ./demo-config/connections.json
	@echo "Demo configuration created in ./demo-config/"
	@echo "To test with demo config, set UPDUCK_CONFIG_DIR=./demo-config"

run-tower: build demo
	@echo "Starting demo tower server..."
	UPDUCK_CONFIG_DIR=./demo-config ./upduck server --type=tower --port=8080

run-server: build demo
	@echo "Starting demo server..."
	UPDUCK_CONFIG_DIR=./demo-config ./upduck server --type=server --port=8081
