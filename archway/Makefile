GO ?= go
WIRE_PACKAGE := github.com/google/wire/cmd/wire@v0.7.0
WIRE_BIN := $(or $(shell command -v wire 2>/dev/null),$(shell $(GO) env GOPATH)/bin/wire)

ROOT_DIR := $(CURDIR)
BUILD_DIR := $(ROOT_DIR)/.build
GATEWAY_DIR := $(ROOT_DIR)/gateway
GATEWAY_CONTROLLER_DIR := $(ROOT_DIR)/gateway-controller
PROTOBUF_DIR := $(ROOT_DIR)/../protobuf

.PHONY: all wire build clean wire-bin wire-gateway wire-gateway-controller build-gateway build-gateway-controller proto-upstream-lookup

all: build

wire: wire-gateway wire-gateway-controller

build: wire build-gateway build-gateway-controller

clean:
	rm -rf $(BUILD_DIR)

wire-bin:
	@if command -v wire >/dev/null 2>&1; then \
		echo "wire already installed"; \
	else \
		echo "installing wire"; \
		$(GO) install $(WIRE_PACKAGE); \
	fi

wire-gateway: wire-bin
	cd $(GATEWAY_DIR)/cmd && $(WIRE_BIN)

wire-gateway-controller: wire-bin
	cd $(GATEWAY_CONTROLLER_DIR)/cmd && $(WIRE_BIN)

build-gateway: wire-gateway
	mkdir -p $(BUILD_DIR)
	cd $(GATEWAY_DIR) && $(GO) build -o $(BUILD_DIR)/gateway ./cmd

build-gateway-controller: wire-gateway-controller
	mkdir -p $(BUILD_DIR)
	cd $(GATEWAY_CONTROLLER_DIR) && $(GO) build -o $(BUILD_DIR)/gateway-controller ./cmd

proto-upstream-lookup:
	cd $(PROTOBUF_DIR) && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative upstream_lookup.proto
