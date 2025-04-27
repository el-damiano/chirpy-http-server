package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) ServeMetrics(writer http.ResponseWriter, request *http.Request) {
	_ = request
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	fmt.Fprintf(writer, "Hits: %d", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) ServeMetricsReset(writer http.ResponseWriter, request *http.Request) {
	_ = request
	cfg.fileserverHits.Store(0)
	writer.WriteHeader(http.StatusOK)
	fmt.Fprintln(writer, "Fileserver hits reset")
}

func ServeReady(writer http.ResponseWriter, request *http.Request) {
	_ = request
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {
	const port = "8080"
	const filePath = "."
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	dir := http.Dir(filePath)

	serveMux := http.NewServeMux()
	serveFile := http.StripPrefix("/app", http.FileServer(dir))
	serveMux.Handle("/app/", apiCfg.metricsMiddleware(serveFile))
	serveMux.HandleFunc("GET /healthz", ServeReady)
	serveMux.HandleFunc("GET /metrics", apiCfg.ServeMetrics)
	serveMux.HandleFunc("POST /reset", apiCfg.ServeMetricsReset)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filePath, port)
	log.Fatal(server.ListenAndServe())
}
