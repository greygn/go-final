# Go Forum Microservices Project

This project consists of two microservices: an authentication service and a forum service with real-time chat functionality.

## Project Structure

```
.
├── auth-service/         # Authentication microservice
├── forum-service/       # Forum and chat microservice
├── protos/             # Protocol buffer definitions
├── migrations/         # Database migrations
├── docker-compose.yml  # Docker compose configuration
└── Makefile           # Build and run commands
```

## Prerequisites

- Go 1.19 or later
- PostgreSQL 13 or later
- Protocol Buffers compiler
- Docker and Docker Compose (optional)
- golang-migrate

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up the databases:
```bash
# Create databases
createdb -U postgres auth_db
createdb -U postgres forum_db

# Run migrations
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/auth_db?sslmode=disable" up
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/forum_db?sslmode=disable" up
```

3. Generate Protocol Buffer files:
```bash
cd protos
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto
```

## Running the Services

### Using Docker Compose

```bash
docker-compose up
```

### Manual Start

1. Start Auth Service:
```bash
cd auth-service
go run cmd/main.go
```

2. Start Forum Service:
```bash
cd forum-service
go run cmd/main.go
```

## API Documentation

API documentation is available at:
- Auth Service: http://localhost:8080/swagger/index.html
- Forum Service: http://localhost:8081/swagger/index.html

## Testing

Run tests with coverage:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Environment Variables

### Auth Service
- `DB_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT tokens
- `ACCESS_TOKEN_TTL` - Access token time to live
- `REFRESH_TOKEN_TTL` - Refresh token time to live
- `GRPC_PORT` - gRPC server port

### Forum Service
- `DB_URL` - PostgreSQL connection string
- `AUTH_SERVICE_ADDR` - Auth service gRPC address
- `HTTP_PORT` - HTTP server port
- `MESSAGE_TTL` - Chat message time to live (default 20s)

## Features

1. Authentication Service:
   - User registration and login
   - JWT-based authentication (access and refresh tokens)
   - Token management

2. Forum Service:
   - Public chat room
   - WebSocket-based real-time messaging
   - Automatic message cleanup (20s TTL)
   - Read access for all users
   - Write access for authenticated users only 