FILES := $(shell go list ./...)

default: build

build:
	@(go build -o ./bin/helgart main.go)

run: build
	@(./bin/helgart)

lint:
	@(golint --set_exit_status ${FILES})

.PHONY: build run lint