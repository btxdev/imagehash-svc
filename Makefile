.PHONY: build test lint run gen-proto gen-docs

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

run: build
	./bin/server

gen-proto:
	export PATH="$$PATH:$$(go env GOPATH)/bin" && \
	cd imagehash && \
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative imagehash.proto

gen-docs:
	protoc --openapiv2_out=api --openapiv2_opt=logtostderr=true imagehash/imagehash.proto