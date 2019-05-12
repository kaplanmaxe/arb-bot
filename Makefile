build:
	@(go build -o ./bin/helgart main.go)

run: build
	@(./bin/helgart)