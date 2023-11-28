.PHONY: build run build-proto

build:
	docker build \
		-f Dockerfile.server \
		-t nehalem90/mocks3_server .

run:
	docker run -p 50051:50051 nehalem90/mocks3_server

build-proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/file_service.proto