#!/bin/sh

protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf:. --gofast_out=plugins=grpc:. ./protobuf/protobuf.proto