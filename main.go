package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server listening on port 10000...")

	// Run the server in a goroutine
	go func() {
		err := http.ListenAndServe("0.0.0.0:10000", nil)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()

	// Listen for interrupt signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Shutting down server...")
}
