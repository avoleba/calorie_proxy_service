.PHONY: build run test docker-up docker-down clean

build:
	go build -o bin/proxy cmd/proxy/main.go

run:
	go run cmd/proxy/main.go

test:
	go test -v ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down

clean:
	rm -rf bin/
	docker compose down -v

lint:
	golangci-lint run

swagger:
	swag init -g cmd/proxy/main.go