# Variables
BINARY_NAME := LeAggLayer
OUTPUT_DIR := build
SOURCE_DIR := ./
GO_FILES := $(wildcard $(SOURCE_DIR)/*.go)

# Default target
.PHONY: all
all: build

# Build target
.PHONY: build
build:
	@echo "Building the binary..."
	mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) $(SOURCE_DIR)

# Run target
.PHONY: run
run: build
	@echo "Running the binary..."
	./$(OUTPUT_DIR)/$(BINARY_NAME)

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf $(OUTPUT_DIR)

# Cross-compile for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux $(SOURCE_DIR)

# Cross-compile for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(BINARY_NAME).exe $(SOURCE_DIR)
