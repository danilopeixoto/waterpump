.PHONY: all

SOURCE_DIR := waterpump/
BUILD_DIR := build/

all: dependencies build

dependencies:
	go -C $(SOURCE_DIR) mod tidy

format:
	go -C $(SOURCE_DIR) fmt ./...

build:
	mkdir -p $(BUILD_DIR)
	go -C $(SOURCE_DIR) build -o ../$(BUILD_DIR)/waterpump ./cmd/waterpump

clean:
	rm -rf $(BUILD_DIR)
