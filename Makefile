.PHONY: proto test migrate-up migrate-down run-auth run-forum

# Protocol buffer compilation
proto:
	cd protos && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

# Database migrations
migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/auth_db?sslmode=disable" up
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/forum_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/auth_db?sslmode=disable" down
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/forum_db?sslmode=disable" down

# Run services
run-auth:
	cd auth-service && go run cmd/main.go

run-forum:
	cd forum-service && go run cmd/main.go

# Testing
test:
	go test ./... -v -cover

# Docker
docker-build:
	docker-compose build

docker-up:
	docker-compose up

docker-down:
	docker-compose down

# Generate Swagger documentation
swagger:
	cd auth-service && swag init -g cmd/main.go
	cd forum-service && swag init -g cmd/main.go

# Additional commands
lint:
	cd auth-service && golangci-lint run
	cd forum-service && golangci-lint run

docs:
	cd auth-service && swag init -g cmd/main.go
	cd forum-service && swag init -g cmd/main.go

logs:
	docker-compose logs -f 