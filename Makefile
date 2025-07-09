PROTO_SRC = api/protos
OUT_DIR = pkg/grpc

PROTO_FILES := $(wildcard $(PROTO_SRC)/*.proto)

gen-proto:
	protoc \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)	
	mv $(OUT_DIR)/api/protos/*.pb.go $(OUT_DIR)/
	rm -rf $(OUT_DIR)/api


gen-mock:	
	mockgen -source=$(file) \
		-destination=$(dir $(file))$(notdir $(basename $(file)))_mock.go \
		-package=$(shell basename $(dir $(file)))

test:
	go test ./... -cover