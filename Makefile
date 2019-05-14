FILES := $(shell go list ./...)

default: build

build:
	@(go build -o ./bin/helgart main.go)

run: build
	@(./bin/helgart)

lint:
	@(golint --set_exit_status ${FILES})

unit:
	@(go test -cover ./...)

test: lint unit

.PHONY: build run lint unit test