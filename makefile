.PHONY: proto clean

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

# Proto related variables
PROTO_DIR=proto
PROTO_FILES=$(wildcard $(PROTO_DIR)/*.proto)

# Tools
PROTOC=protoc
PROTOC_GEN_GO=$(GOBIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC=$(GOBIN)/protoc-gen-go-grpc

# Install protoc plugins
install-proto-tools:
	@echo "Installing protoc-gen-go..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "Installing protoc-gen-go-grpc..."
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate proto files
proto: install-proto-tools
	@echo "Generating protobuf files..."
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f $(PROTO_DIR)/*.pb.go

# Build all
build: proto
	@echo "Building..."
	go build -o $(GOBIN)/server cmd/server/rest/main.go
	go build -o $(GOBIN)/grpc-server cmd/server/grpc/main.go
	go build -o $(GOBIN)/client cmd/client/main.go

# Run server
run-server: build
	@echo "Running server..."
	$(GOBIN)/server

# Run gRPC server
run-grpc-server: build
	@echo "Running gRPC server..."
	$(GOBIN)/grpc-server

# Run client
run-client: build
	@echo "Running client..."
	$(GOBIN)/client

# Help
help:
	@echo "Available targets:"
	@echo "  make proto          - Generate protobuf files"
	@echo "  make clean         - Clean generated files"
	@echo "  make build         - Build all binaries"
	@echo "  make run-server    - Run the server"
	@echo "  make run-grpc-server - Run the gRPC server"
	@echo "  make run-client    - Run the client"
	@echo "  make help          - Show this help"
