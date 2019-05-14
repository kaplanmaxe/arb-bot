package mock

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
)

var upgrader = websocket.Upgrader{}

type echoServer struct {
	upgrader websocket.Upgrader
	Done     chan struct{}
}

func (s *echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	s.Done = make(chan struct{})
	defer close(s.Done)

	if r.URL.Path != "/ws" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var conn *websocket.Conn
	conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Fatalf("Error writing message: %s", err)
		}
	}
}

func NewWebsocketServer() *websocket.Dialer {
	return wstest.NewDialer(&echoServer{})
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}
