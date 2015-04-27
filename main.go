package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader    = websocket.Upgrader{}
	connections = make(map[*websocket.Conn]bool)
)

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func broadcast(msg []byte) {
	log.Printf("broadcasting: %s", string(msg))
	for conn, _ := range connections {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade request to ws connection, err: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	connections[conn] = true
	defer func() {
		delete(connections, conn)
		conn.Close()
	}()
	if err := conn.WriteMessage(websocket.TextMessage, []byte("connection stablished")); err != nil {
		log.Printf("failed to write message, err: %v", err)
		return
	}
	for {
		_, msg, err := conn.ReadMessage()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("failed to read message, err: %v", err)
			return
		}
		broadcast(msg)
	}
}

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/ws", WSHandler)
	http.ListenAndServe(":8080", nil)
}
