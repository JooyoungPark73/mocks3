.PHONY: build run proto

build:
	docker build \
		--build-arg FUNC_PORT=30000 \
		--build-arg FUNC_VERBOSE_ENV=debug \
		-f Dockerfile.server \
		-t nehalem90/mocks3_server .
		docker push nehalem90/mocks3_server:latest

run:
	docker run -p 30000:30000 nehalem90/mocks3_server

proto:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/file_service.proto