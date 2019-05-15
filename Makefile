FILES := $(shell go list ./...)

default: build

build:
	@(go build -o ./bin/helgart main.go)

run: build
	@(./bin/helgart)

run-race: build
	go run -race main.go

lint:
	@(golint --set_exit_status ${FILES})

unit:
	@(go test -cover ./...)

unit-race:
	@(go test -race -cover ./...)

test: lint unit unit-race

.PHONY: build run run-race lint unit unit-race test