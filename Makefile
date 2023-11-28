.PHONY: build run proto

build:
	docker build \
		--build-arg FUNC_PORT=50051 \
		-f Dockerfile.server \
		-t nehalem90/mocks3_server .
		docker push nehalem90/mocks3_server:latest

run:
	docker run -p 50051:50051 nehalem90/mocks3_server

proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/file_service.proto