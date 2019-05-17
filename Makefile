FILES := $(shell go list ./...)

default: build

build:
	@(go build -o ./bin/helgart main.go)

run: build
	@(./bin/helgart)

run-race: build
	go run -race main.go

up:
	@(docker-compose up -d)

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

.PHONY: build run run-race up down lint unit unit-race test mysql