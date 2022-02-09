#!/bin/bash
# protoc --plugin=protoc-gen-mqant=protoc-gen-mqant --plugin=protoc-gen-go=/Users/wangxinyu/go/bin/protoc-gen-go --proto_path=proto --mqant_out=./ --go_out=./  proto/examples1/*greeter_1.proto --experimental_allow_proto3_optional --go_out=paths=source_relative:.
protoc --plugin=protoc-gen-go=/Users/wangxinyu/go/bin/protoc-gen-go   --go_out=./  ./greeter/*.proto --experimental_allow_proto3_optional
protoc --plugin=protoc-gen-mqant=/Users/wangxinyu/go/bin/protoc-gen-mqant  --proto_path=.  --mqant_out=./   ./greeter/*.proto ./greeter/*.proto --experimental_allow_proto3_optional