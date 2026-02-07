APP_NAME = gateway
BIN_DIR = bin
MAIN = cmd/main.go

# 默认参数（可被命令行 ARGS 覆盖）
ARGS ?= -config=gateway.yaml -output_console=true

.PHONY: all build run test clean lint docker

all: build

build:
	go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN)

run:
	go run $(MAIN) $(ARGS)

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)

fmt:
	go fmt ./...

lint:
	golangci-lint run

docker:
	docker build -t $(APP_NAME):latest .