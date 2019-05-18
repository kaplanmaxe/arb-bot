FILES := $(shell go list ./...)
BROKER_BIN = helgart-broker
PROJECT_ROOT = $(shell pwd)

default: build

build-broker:
	@(go build -v -o ./bin/${BROKER_BIN} ./cmd/broker/main.go)

run-broker: build-broker
	./bin/${BROKER_BIN} --config ${PROJECT_ROOT}/.config.yml

run-race: build
	go run -race main.go

up:
	@(docker-compose up -d)
	@(docker logs --follow helgart_broker)

down:
	@(docker-compose down)

lint:
	@(golint --set_exit_status ${FILES})

unit:
	@(go test -cover ./...)

unit-race:
	@(go test -race -cover ./...)

test: lint unit unit-race

mysql:
	@(mysql -u root -h 0.0.0.0 -P 33104 -p)

.PHONY: build-broker run-broker run run-race up down lint unit unit-race test mysql