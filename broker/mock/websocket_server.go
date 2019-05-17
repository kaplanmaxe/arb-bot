package mock

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
)

var upgrader = websocket.Upgrader{}

type echoServer struct {
	upgrader   websocket.Upgrader
	Done       chan struct{}
	ignoreFunc func(msg []byte) bool
}

// ServeHTTP is the handler func to serve websocket traffic
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
		if s.ignoreFunc(message) == false {
			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Fatalf("Error writing message: %s", err)
			}
		}

	}
}

// NewWebsocketServer returns a new websocket server
func NewWebsocketServer(ignoreFunc func(msg []byte) bool) *websocket.Dialer {
	return wstest.NewDialer(&echoServer{
		ignoreFunc: ignoreFunc,
	})
}
