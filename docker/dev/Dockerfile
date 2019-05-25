FROM golang:1.12.4-alpine

RUN apk add --no-cache ca-certificates git make

RUN go get github.com/githubnemo/CompileDaemon

WORKDIR /go/src/github.com/kaplanmaxe/helgart

COPY . .

EXPOSE 8080/tcp
EXPOSE 8080/udp

# RUN make install

CMD CompileDaemon -build="go build -o ./bin/broker ./broker/main.go" -directory="." -command="./bin/broker --config /go/src/github.com/kaplanmaxe/helgart/.config.docker.yml"