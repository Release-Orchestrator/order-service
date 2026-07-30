.PHONY: build test lint run docker-build tidy

BINARY_NAME=order-service
DOCKER_IMAGE=ghcr.io/release-orchestrator/order-service
VERSION?=dev

build:
	go build -o bin/$(BINARY_NAME) ./cmd

test:
	go test -v ./...

lint:
	golangci-lint run ./...

run: build
	DATABASE_URL="postgres://postgres:postgres@localhost:5434/order_db?sslmode=disable" USER_SERVICE_URL="http://localhost:8081" PAYMENT_SERVICE_URL="http://localhost:8082" ./bin/$(BINARY_NAME)

docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

tidy:
	go mod tidy
