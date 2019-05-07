build:
	@(go build -o ./bin/cw-websocket main.go)

run: build
	@(./bin/cw-websocket)