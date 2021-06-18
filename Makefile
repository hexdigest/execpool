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
	rm -f ./profile.cov
	GOGC=off go test -race ./... -count 1 -v -coverprofile=profile.cov

.PHONY:
benchmark: lint
	GOGC=off go test -bench=. ./... -benchtime=100x -v -count 1

.PHONY:
clean:
	rm -f ./bin/*
	rm -f ./profile.cov
	rm -f ./execpool.test

./profile.cov: test

.PHONY:
show-coverage: profile.cov
	go tool cover -html=profile.cov

.PHONY:
tidy:
	go mod tidy
