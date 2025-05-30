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
	tokenSecret := os.Getenv("SECRET")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	const port = "8080"
	const filePath = "."
	apiCfg := &apiConfig{
		dbQueries:   database.New(db),
		platform:    platform,
		tokenSecret: tokenSecret,
	}
	dir := http.Dir(filePath)

	serveMux := http.NewServeMux()
	fileHandler := http.StripPrefix("/app", http.FileServer(dir))
	serveMux.Handle("/app/", apiCfg.metricsMiddleware(fileHandler))

	serveMux.HandleFunc("GET /api/healthz", readyHandler)

	serveMux.HandleFunc("POST /api/users", apiCfg.userCreateHandler)
	serveMux.HandleFunc("PUT /api/users", apiCfg.userUpdateHandler)
	serveMux.HandleFunc("POST /api/login", apiCfg.userLoginHandler)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.tokenRefreshHandler)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.tokenRevokeHandler)

	serveMux.HandleFunc("POST /api/chirps", apiCfg.chirpCreateHandler)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.chirpsHandler)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.chirpsHandler)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.chirpsDeleteHandler)

	serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.metricsResetHandler)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filePath, port)
	log.Fatal(server.ListenAndServe())
}
