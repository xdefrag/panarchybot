GITHEAD:=$(shell git rev-parse HEAD)
CMD:=./cmd/panarchybot/...
OUTPUT:=./dist/panarchybot
LDFLAGS:=-ldflags="-X 'main.Commit=${GITHEAD}'"

default: build

build:
	sqlc generate
	go generate ./panarchybot.go
	go build ${LDFLAGS} -o ${OUTPUT} ${CMD}

.PHONY: run
run: build
	${OUTPUT}

.PHONY: test
test: build
	go test -v -count=1 ./...

.PHONY: migrate-status
migrate-status:
	goose -dir migrations postgres "dbname=mlm sslmode=disable" status

.PHONY: migrate-generate
migrate-generate:
	goose -dir migrations postgres "dbname=mlm sslmode=disable" create ${name} sql

.PHONY: migrate-up
migrate-up:
	goose -dir migrations postgres "dbname=mlm sslmode=disable" up