
protoc:
	@protoc -I ./proto/docs --go-grpc_out=./proto --go-grpc_opt=paths=source_relative --go_out=./proto --go_opt=paths=source_relative ./proto/docs/*.proto
