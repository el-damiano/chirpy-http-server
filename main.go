package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	serveMux := http.NewServeMux()
	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
