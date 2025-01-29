.PHONY: all clean

# Go package path
GO_PKG := github.com/galxe/spotted-network/proto

# Protocol buffer files
PROTO_FILES := $(wildcard *.proto)

# Output files
PB_GO_FILES := $(PROTO_FILES:.proto=.pb.go)
GRPC_GO_FILES := $(PROTO_FILES:.proto=_grpc.pb.go)

all: $(PB_GO_FILES) $(GRPC_GO_FILES)

%.pb.go %_grpc.pb.go: %.proto
	protoc \
		--go_out=. \
		--go_opt=module=$(GO_PKG) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(GO_PKG) \
		$<

clean:
	rm -f *.pb.go *_grpc.pb.go


