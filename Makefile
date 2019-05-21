FILES := $(shell go list ./...)
BROKER_BIN = broker
PROJECT_ROOT = $(shell pwd)

default: build

# BUILD
build-broker:
	@(go build -v -o ./bin/${BROKER_BIN} ./broker/main.go)

build: build-broker

# install
install:
	@(go install github.com/kaplanmaxe/helgart/broker)

# RUN
dev-up:
	@(docker-compose -f docker/dev/docker-compose.dev.yml up -d)
	@(docker logs --follow helgart_broker_dev)

dev-down:
	@(docker-compose -f docker/dev/docker-compose.dev.yml down)

run-broker: build-broker
	./bin/${BROKER_BIN}

run-version: build-broker
	./bin/${BROKER_BIN} --version

run: run-broker

up:
	@(docker-compose up --build -d)

broker-logs:
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

.PHONY: build build-broker dev-up dev-down run run-version run-broker run run-race up broker-logs down lint unit unit-race test mysql