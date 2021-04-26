#!/bin/sh

rm -rf api gen
mkdir api gen

protoc -I/usr/local/include -I. --proto_path=. \
--go_out=gen \
--go-grpc_out=gen \
--message-validator_out=gen \
--grpc-gateway_out=logtostderr=true:gen \
--openapiv2_out=logtostderr=true:api \
*.proto