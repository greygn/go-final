.PHONY: gen clean

gen:
	mkdir -p gen/go
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/forum.proto proto/auth.proto

clean:
	rm -rf gen 