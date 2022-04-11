.PHONY: all test

all: test build_receiver build_plugins

build_receiver:
	go build -o bin/receiver ./cli/receiver

build_plugins:
	go build -buildmode=plugin -o bin/clickhouse.so ./libs/store/clickhouse/clickhouse.go

test:
	go test ./...