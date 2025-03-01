build_proto:
	protoc --go_out=./proto --go_opt=paths=source_relative \
	       --go-grpc_out=./proto --go-grpc_opt=paths=source_relative \
	       proto/user.proto
build_grpc_ui:
	grpcui -plaintext localhost:9001