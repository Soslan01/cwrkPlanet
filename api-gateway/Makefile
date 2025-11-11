SHELL := /bin/bash

APP_NAME := api-gateway
BIN_DIR  := bin
BIN      := $(BIN_DIR)/$(APP_NAME)

GO := go

.PHONY: all
all: run

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) ./cmd

.PHONY: run
run: build
	CONFIG_PATH=internal/config/config.yaml ./$(BIN)

.PHONY: test
test:
	$(GO) test ./...

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
