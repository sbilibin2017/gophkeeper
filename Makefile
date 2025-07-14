PROTO_SRC = api/grpc
OUT_DIR = pkg/grpc

PROTO_FILES := $(wildcard $(PROTO_SRC)/*.proto)

gen-proto:
	protoc \
		--proto_path=$(PROTO_SRC) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

gen-mock:
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test ./... -cover

build-clients:
	GOOS=linux GOARCH=amd64 go build -o builds/gophkeeper-client-linux-amd64 ./cmd/client
	GOOS=darwin GOARCH=amd64 go build -o builds/gophkeeper-client-macos-amd64 ./cmd/client
	GOOS=windows GOARCH=amd64 go build -o builds/gophkeeper-client-windows-amd64.exe ./cmd/client
