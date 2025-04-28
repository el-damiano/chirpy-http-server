package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const filePath = "."
	apiCfg := &apiConfig{}
	dir := http.Dir(filePath)

	serveMux := http.NewServeMux()
	fileHandler := http.StripPrefix("/app", http.FileServer(dir))
	serveMux.Handle("/app/", apiCfg.metricsMiddleware(fileHandler))
	serveMux.HandleFunc("GET /api/healthz", readyHandler)
	serveMux.HandleFunc("POST /api/validate_chirp", chirpValidateHandler)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.metricsResetHandler)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filePath, port)
	log.Fatal(server.ListenAndServe())
}
