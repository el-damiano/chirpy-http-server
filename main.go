package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/el-damiano/bootdev-http-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading env files")
		return
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	const port = "8080"
	const filePath = "."
	apiCfg := &apiConfig{
		dbQueries: database.New(db),
		platform:  platform,
	}
	dir := http.Dir(filePath)

	serveMux := http.NewServeMux()
	fileHandler := http.StripPrefix("/app", http.FileServer(dir))
	serveMux.Handle("/app/", apiCfg.metricsMiddleware(fileHandler))
	serveMux.HandleFunc("GET /api/healthz", readyHandler)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.chirpCreateHandler)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.chirpsHandler)
	serveMux.HandleFunc("POST /api/users", apiCfg.userCreateHandler)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.metricsResetHandler)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filePath, port)
	log.Fatal(server.ListenAndServe())
}
