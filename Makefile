# Paths
src := api/grpc
dst := pkg/grpc

# Generate Go code from .proto files
gen-proto:
	@mkdir -p $(dst)
	@for protofile in $(wildcard $(src)/*.proto); do \
		echo "Generating $$protofile into $(dst)"; \
		protoc \
			--proto_path=$(src) \
			--go_out=$(dst) --go_opt=paths=source_relative \
			--go-grpc_out=$(dst) --go-grpc_opt=paths=source_relative \
			$$protofile; \
	done

# Generate mocks
gen-mock:
	mockgen -source=$(file) \
		-destination=$(dir $(file))/$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

# Run tests with coverage
test:
	go test ./... -cover

# Build client binaries for major OS/ARCH combos
build-clients:
	GOOS=linux GOARCH=amd64 go build -o builds/gophkeeper-client-linux-amd64 ./cmd/client
	GOOS=darwin GOARCH=amd64 go build -o builds/gophkeeper-client-macos-amd64 ./cmd/client
	GOOS=windows GOARCH=amd64 go build -o builds/gophkeeper-client-windows-amd64.exe ./cmd/client
