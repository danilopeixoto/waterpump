.PHONY: all

SOURCE_DIR := waterpump/
BUILD_DIR := build/

all: dependencies build

dependencies:
	go -C $(SOURCE_DIR) mod tidy

build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go -C $(SOURCE_DIR) build -o ../$(BUILD_DIR)/waterpump

clean:
	rm -rf $(BUILD_DIR)
