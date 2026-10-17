BINARY_NAME=bun
ALIAS_NAME=coffee-bun
BIN_DIR=bin
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: all build install test clean lint

all: build

build:
	@mkdir -p $(BIN_DIR)
	@rm -f $(BIN_DIR)/$(BINARY_NAME) $(BIN_DIR)/$(ALIAS_NAME)
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/bun
	@which codesign >/dev/null 2>&1 && codesign -s - -f $(BIN_DIR)/$(BINARY_NAME) 2>/dev/null || true
	@ln -sf $(BINARY_NAME) $(BIN_DIR)/$(ALIAS_NAME)
	@echo "Binaries built at $(BIN_DIR)/$(BINARY_NAME) and $(BIN_DIR)/$(ALIAS_NAME)"

install: build
	@mkdir -p $(INSTALL_DIR)
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(ALIAS_NAME)
	install -m 755 $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@which codesign >/dev/null 2>&1 && codesign -s - -f $(INSTALL_DIR)/$(BINARY_NAME) 2>/dev/null || true
	@ln -sf $(INSTALL_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(ALIAS_NAME)
	@echo "Installed $(BINARY_NAME) and $(ALIAS_NAME) to $(INSTALL_DIR)"

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
