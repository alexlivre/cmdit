.PHONY: build run test lint clean build-all

GO := go
BIN := bin/cmdit.exe
BIN_LINUX := bin/cmdit-linux-amd64
BIN_WIN := bin/cmdit-windows-amd64.exe
BIN_MAC := bin/cmdit-darwin-amd64
BIN_MAC_ARM := bin/cmdit-darwin-arm64

build:
	$(GO) build -o $(BIN) ./cmd/cmdit

run: build
	$(BIN)

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

build-all:
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BIN_LINUX) ./cmd/cmdit
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BIN_WIN) ./cmd/cmdit
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BIN_MAC) ./cmd/cmdit
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BIN_MAC_ARM) ./cmd/cmdit

clean:
	rm -rf bin/
