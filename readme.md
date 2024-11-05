## Generate GRPC

```sh
protoc \
--plugin=protoc-gen-go_grpc=/usr/bin/protoc-gen-go-grpc \
--go_out=. \
--go_opt=paths=source_relative \
--go_grpc_out=. \
--go_grpc_opt=paths=source_relative \
consensus/consensus.proto
```

Note:
--Plugin er optional