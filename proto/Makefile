.PHONY: proto

all: proto

proto:
	mkdir -p gen
	protoc -I . \
		-I ./protoc-gen-validate \
		--go_out ./gen --go_opt paths=source_relative \
		--go-grpc_out ./gen --go-grpc_opt  paths=source_relative \
		--validate_out="lang=go,paths=source_relative:./gen" \
		./api.proto

install: install_go_dependencies

install_go_dependencies:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
