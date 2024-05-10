package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	// Create a new Upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow all connections
			return true
		},
	}

	// Define WebSocket upgrade handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, upgrader)
	})

	// Start server
	address := "0.0.0.0:10000" // Listen on all interfaces on port 8080
	log.Printf("Server started and listening on %s\n", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v\n", err)
		return
	}
	defer conn.Close()

	// Send confirmation message
	if err := conn.WriteMessage(websocket.TextMessage, []byte("Hello, world!")); err != nil {
		log.Printf("Failed to send confirmation message: %v\n", err)
		return
	}

	// Keep connection open
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			break
		}
	}
}
