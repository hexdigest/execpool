export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)
export GO111MODULE := on

./bin/minimock:
	go install github.com/gojuno/minimock/v3/cmd/minimock

./bin/gowrap:
	go install github.com/hexdigest/gowrap/cmd/gowrap

./bin/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY:
tools: ./bin/minimock ./bin/golangci-lint

.PHONY:
generate: tools
	go generate ./...

.PHONY:
lint: ./bin/golangci-lint
	golangci-lint run --enable=goimports ./...

.PHONY:
test: lint
	GOGC=off go test -race ./... -count 1 -v

.PHONY:
benchmark: lint
	GOGC=off go test -bench=. ./... -benchtime=100x -v -count 1

.PHONY:
clean:
	rm -f ./bin/*

.PHONY:
coverage:
	rm -f coverage.out
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY:
tidy:
	go mod tidy
