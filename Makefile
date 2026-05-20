.PHONY: build build-all package clean

BINARY_NAME=log-compare
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS=-ldflags "-s -w \
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME)"

build:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build success: $(BUILD_DIR)/$(BINARY_NAME)"

build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_amd64 .
	@echo "Build success: $(BUILD_DIR)/$(BINARY_NAME)_amd64"
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_arm64 .
	@echo "Build success: $(BUILD_DIR)/$(BINARY_NAME)_arm64"

package: build-all
	@mkdir -p $(BUILD_DIR)/pkg/$(BINARY_NAME)
	@cp $(BUILD_DIR)/$(BINARY_NAME)_amd64 $(BUILD_DIR)/pkg/$(BINARY_NAME)/$(BINARY_NAME)_amd64
	@cp $(BUILD_DIR)/$(BINARY_NAME)_arm64 $(BUILD_DIR)/pkg/$(BINARY_NAME)/$(BINARY_NAME)_arm64
	@cp -r conf $(BUILD_DIR)/pkg/$(BINARY_NAME)/ 2>/dev/null || true
	@tar -czf $(BUILD_DIR)/$(BINARY_NAME).tar.gz -C $(BUILD_DIR)/pkg $(BINARY_NAME)
	@echo "Package success: $(BUILD_DIR)/$(BINARY_NAME).tar.gz"
	@rm -rf $(BUILD_DIR)/pkg

clean:
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build artifacts"
