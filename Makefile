FILES := $(shell go list ./...)
BROKER_BIN = helgart-broker
PROJECT_ROOT = $(shell pwd)

default: build

# BUILD
build-broker:
	@(go build -v -o ./bin/${BROKER_BIN} ./cmd/broker/main.go)

build: build-broker

# install
install:
	@(go install github.com/kaplanmaxe/helgart/cmd/broker)

# RUN
run-broker: build-broker
	./bin/${BROKER_BIN} --config ${PROJECT_ROOT}/.config.yml

run-version: build-broker
	./bin/${BROKER_BIN} --version

run: run-broker

up:
	@(docker-compose up -d)
	@(docker logs --follow helgart_broker)

down:
	@(docker-compose down)

# TEST
lint:
	@(golint --set_exit_status ${FILES})

unit:
	@(go test -cover ./...)

unit-race:
	@(go test -race -cover ./...)

test: lint unit unit-race

mysql:
	@(mysql -u root -h 0.0.0.0 -P 33104 -p)

.PHONY: build build-broker run run-version run-broker run run-race up down lint unit unit-race test mysql