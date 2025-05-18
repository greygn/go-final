FROM golang:1.21-alpine AS builder

WORKDIR /src

# Copy the entire project
COPY . .

# Download dependencies
RUN cd forum-service && go mod download

# Build the application
RUN cd forum-service && CGO_ENABLED=0 GOOS=linux go build -o /app/forum-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/forum-service .
COPY --from=builder /src/forum-service/migrations ./migrations

EXPOSE 8081 50052

CMD ["./forum-service"] 