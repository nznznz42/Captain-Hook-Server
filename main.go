package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

type ClientManager struct {
	clients map[string]Client
	mu      sync.Mutex
}

func (cm *ClientManager) AddClient(clientID string, clientHost string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[clientID] = Client{ID: clientID, Host: clientHost}
}

func (cm *ClientManager) GetClient(clientID string) (Client, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	client, ok := cm.clients[clientID]
	return client, ok
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")

	// Simulate sending a request to the Python program after 5 seconds
	go func() {
		time.Sleep(5 * time.Second)
		sendGoodbyeToPython()
	}()
}

func goodbyeHandler(cm *ClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("id")
		if clientID == "" {
			http.Error(w, "Client ID not provided", http.StatusBadRequest)
			return
		}

		client, ok := cm.GetClient(clientID)
		if !ok {
			http.Error(w, "Client not found", http.StatusNotFound)
			return
		}

		// Do something with the client details
		fmt.Fprintf(w, "Goodbye, world to %s at %s!", client.ID, client.Host)
	}
}

func sendGoodbyeToPython() {
	// Simulate sending a request to the Python program
	fmt.Println("Sending 'Goodbye, world!' to Python")
}

func main() {
	cm := &ClientManager{
		clients: make(map[string]Client),
	}

	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/goodbye", goodbyeHandler(cm))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
