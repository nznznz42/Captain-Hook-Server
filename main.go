package main

import (
	"fmt"
	"net/http"
	"time"
)

var (
	clientIP      string
	requestActive bool
)

func handler(w http.ResponseWriter, r *http.Request) {
	clientIP = r.RemoteAddr
	requestActive = true
	fmt.Fprintf(w, "Hello, World!")
}

func sendGoodbye(ip string) {
	fmt.Printf("Sending 'Goodbye, World!' to %s\n", ip)
}

func goodbyeScheduler() {
	for {
		time.Sleep(5 * time.Second)
		if requestActive {
			sendGoodbye(clientIP)
			requestActive = false
		}
	}
}

func main() {
	go goodbyeScheduler() // Start the scheduler goroutine

	http.HandleFunc("/", handler)
	fmt.Println("Server listening on port 10000...")
	err := http.ListenAndServe("0.0.0.0:10000", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
