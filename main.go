package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	//if len(os.Args) > 3 {
	//	fmt.Println(os.Args)
	//	fmt.Println("usage main.go <domain> <port>")
	//	log.Fatalf("invalid parameter count")
	//}
	domain := "whtest.rahulsk.dev"
	port := 10000

	clientsManager := NewManager()
	mux := NewWebHookHandler(clientsManager, domain)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
