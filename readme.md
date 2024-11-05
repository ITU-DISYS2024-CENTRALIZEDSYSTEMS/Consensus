## Generate GRPC

```sh
protoc \
--plugin=protoc-gen-go_grpc=/usr/bin/protoc-gen-go-grpc \
--go_out=consensus \
--go_opt=paths=source_relative \
--go_grpc_out=consensus \
--go_grpc_opt=paths=source_relative \
consensus.proto
```

Note:
--Plugin er optional