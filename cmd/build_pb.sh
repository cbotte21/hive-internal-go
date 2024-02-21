#!/bin/bash

protoc --go_out="$HIVE_DIR"/pb --go_opt=paths=source_relative --go-grpc_out="$HIVE_DIR"/pb --go-grpc_opt=paths=source_relative --proto_path="$PROTO_DIR" hive.proto

protoc --go_out="$HIVE_DIR"/pb --go_opt=paths=source_relative --go-grpc_out="$HIVE_DIR"/pb --go-grpc_opt=paths=source_relative --proto_path="$PROTO_DIR" hive.proto