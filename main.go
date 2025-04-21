package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const filePath = "."

	serveMux := http.NewServeMux()
	dir := http.Dir(filePath)
	serveMux.Handle("/", http.FileServer(dir))

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}
	log.Printf("Serving files from %s on port: %s\n", filePath, port)
	log.Fatal(server.ListenAndServe())
}
