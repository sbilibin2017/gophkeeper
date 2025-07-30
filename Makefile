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

# Run tests and generate coverage profile, filtered for internal packages only
test:
	go test ./... -coverprofile=coverage.out

	@echo "Generating Markdown coverage table file from coverage.out..."
	@echo "### TEST COVERAGE" > coverage.md
	@echo "" >> coverage.md
	@echo "| PACKAGE                                           | COVERAGE |" >> coverage.md
	@echo "|-------------------------------------------------|----------|" >> coverage.md
	@go tool cover -func=coverage.out | \
		grep -v "^total:" | \
		grep "^github.com/sbilibin2017/gophkeeper/internal/" | \
		grep -v "_test.go" | \
		grep -v "_mock.go" | \
		awk '{ \
			pkg=$$1; \
			sub(/:[0-9]+:$$/, "", pkg); \
			cov=substr($$NF, 1, length($$NF)-1); \
			gsub(",", ".", cov); \
			if(cov > max[pkg]) { max[pkg]=cov; } \
		} \
		END { \
			for (p in max) { \
				printf "| %-47s | %8.1f%% |\n", p, max[p]; \
			} \
		}' | sort >> coverage.md

	rm coverage.out

# Build client binaries for major OS/ARCH combos
build-clients:
	GOOS=linux GOARCH=amd64 go build -o builds/client/gophkeeper-client-linux-amd64 ./cmd/client
	GOOS=darwin GOARCH=amd64 go build -o builds/client/gophkeeper-client-macos-amd64 ./cmd/client
	GOOS=windows GOARCH=amd64 go build -o builds/client/gophkeeper-client-windows-amd64.exe ./cmd/client

# Build server binary for linux
build-server:
	GOOS=linux GOARCH=amd64 go build -o builds/server/gophkeeper-server-linux-amd64 ./cmd/server	

# Generate swagger docs from handlers in internal/handlers,
# use cmd/server/main.go as entry point for swag init,
# output to api/http folder
gen-swag:
	swag init -d internal/handlers -g ../../cmd/server/main.go -o api/http