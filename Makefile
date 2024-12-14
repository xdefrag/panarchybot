GITHEAD:=$(shell git rev-parse HEAD)
CMD:=./cmd/panarchybot/...
OUTPUT:=./dist/panarchybot
LDFLAGS:=-ldflags="-s -w -X 'main.Commit=${GITHEAD}'"
POSTGRES_DSN:="dbname=panarchybot sslmode=disable"
GOBIN:=$(CURDIR)/bin
export GOBIN

default: build

bin/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

bin/sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

bin/goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: gen
gen: bin/sqlc
	${GOBIN}/sqlc generate
	go generate ./panarchybot.go

.PHONY: build
build: gen
	go build ${LDFLAGS} -o ${OUTPUT} ${CMD}

.PHONY: run
run: build
	${OUTPUT}

.PHONY: lint
lint: bin/golangci-lint
	${GOBIN}/golangci-lint run

.PHONY: test
test: gen
	go test -v -count=1 ./...

.PHONY: migrate-status
migrate-status: bin/goose
	${GOBIN}/goose -dir migrations postgres ${POSTGRES_DSN} status

.PHONY: migrate-generate
migrate-generate: bin/goose
	${GOBIN}/goose -dir migrations postgres ${POSTGRES_DSN} create ${name} sql

.PHONY: migrate-up
migrate-up: bin/goose
	${GOBIN}/goose -dir migrations postgres ${POSTGRES_DSN} up
