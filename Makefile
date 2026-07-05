.PHONY: build run test lint migrate-up migrate-down docker-build docker-up

build:
	go build -o bin/chat-app ./cmd/server

run:
	go run cmd/server/main.go

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run

migrate-up:
	migrate -path migrations -database "postgres://chat:chat123@localhost:5432/chatapp?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://chat:chat123@localhost:5432/chatapp?sslmode=disable" down 1

docker-build:
	docker build -t chat-app:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/
