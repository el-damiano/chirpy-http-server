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
	serveFile := http.StripPrefix("/app", http.FileServer(dir))
	serveMux.Handle("/app/", apiCfg.metricsMiddleware(serveFile))
	serveMux.HandleFunc("GET /api/healthz", ServeReady)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.ServeMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.ServeMetricsReset)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filePath, port)
	log.Fatal(server.ListenAndServe())
}
