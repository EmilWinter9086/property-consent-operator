package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = ":8080"
	}

	service := ConsentService{Client: NewConsentClient(apiKey)}
	server := &http.Server{Addr: address, Handler: service.Handler()}
	log.Printf("property consent service listening on %s", address)
	log.Fatal(server.ListenAndServe())
}
