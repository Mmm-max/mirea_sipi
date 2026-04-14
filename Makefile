APP_NAME=sipi-api

.PHONY: run build test tidy fmt swag swagger docker-up docker-down

run:
	go run ./cmd/api

build:
	go build -o ./build/$(APP_NAME) ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal ./docs

swag:
	go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g ./cmd/api/main.go -o ./docs

swagger: swag

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down --remove-orphans
