#!/usr/bin/env bash
#go build -o improving-server improving-server.go

 export GOPATH=$GOPATH:$PWD

 echo $GOPATH


# echo "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ludo-server"
# CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o ludo-server main.go

echo "protoc --proto_path=rpc/pb --go_out=rpc/pb --go_opt=paths=source_relative mqant_rpc.proto"

protoc --proto_path=rpc/pb --go_out=rpc/pb --go_opt=paths=source_relative mqant_rpc.proto

echo "protoc --proto_path=gate/base --go_out=gate/base --go_opt=paths=source_relative session.proto"

protoc --proto_path=gate/base --go_out=gate/base --go_opt=paths=source_relative session.proto

echo "protoc --proto_path=httpgateway/proto --go_out=httpgateway/proto --go_opt=paths=source_relative api.proto"

protoc --proto_path=httpgateway/proto --go_out=httpgateway/proto --go_opt=paths=source_relative api.proto