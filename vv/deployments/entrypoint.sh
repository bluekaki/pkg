#!/bin/sh

rm -rf api gen
mkdir api gen

protoc -I/usr/local/include -I. --proto_path=. \
--go_out=gen \
--go-grpc_out=gen \
--message-validator_out=gen \
--grpc-gateway_out=logtostderr=true:gen \
--openapiv2_out=json_names_for_fields=false,logtostderr=true:api \
*.proto