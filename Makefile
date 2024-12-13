GITHEAD:=$(shell git rev-parse HEAD)
CMD:=./cmd/panarchybot/...
OUTPUT:=./dist/panarchybot
LDFLAGS:=-ldflags="-X 'main.Commit=${GITHEAD}'"
POSTGRES_DSN:="dbname=panarchybot sslmode=disable"

default: build

.PHONY: gen
gen:
	sqlc generate
	go generate ./panarchybot.go

.PHONY: build
build: gen
	go build ${LDFLAGS} -o ${OUTPUT} ${CMD}

.PHONY: run
run: build
	${OUTPUT}

.PHONY: lint
lint: gen
	golangci-lint run

.PHONY: test
test: gen
	go test -v -count=1 ./...

.PHONY: migrate-status
migrate-status:
	goose -dir migrations postgres ${POSTGRES_DSN} status

.PHONY: migrate-generate
migrate-generate:
	goose -dir migrations postgres ${POSTGRES_DSN} create ${name} sql

.PHONY: migrate-up
migrate-up:
	goose -dir migrations postgres ${POSTGRES_DSN} up
