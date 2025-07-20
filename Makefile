gen-proto:
	@for protofile in $(wildcard $(src)/*.proto); do \
		pkg=$$(grep "^package " $$protofile | awk '{print $$2}' | sed 's/;//'); \
		echo "Generating $$protofile into $(dst)/$$pkg"; \
		mkdir -p $(dst)/$$pkg; \
		protoc \
			--proto_path=$(src) \
			--go_out=$(dst)/$$pkg --go_opt=paths=source_relative \
			--go-grpc_out=$(dst)/$$pkg --go-grpc_opt=paths=source_relative \
			$$protofile; \
	done


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
