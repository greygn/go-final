module github.com/greygn/forum-service

go 1.24.3

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.0
	github.com/lib/pq v1.10.9
	github.com/rs/zerolog v1.29.1
	go.uber.org/zap v1.27.0
	protos v0.0.0
)

replace protos => ../protos

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

require (
	github.com/golang-migrate/migrate/v4 v4.18.3
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	google.golang.org/grpc v1.72.1
)
