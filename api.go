package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/el-damiano/bootdev-http-server/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	platform       string
	dbQueries      *database.Queries
	fileserverHits atomic.Int32
}

type Chirp struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) chirpCreateHandler(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()

	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding chirp: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirpValidated := len(chirp.Body) <= 140
	if !chirpValidated {
		writer.WriteHeader(http.StatusBadRequest)
	}

	chirpClean := chirpProfanityFilter(chirp)
	data, err := json.Marshal(chirpClean)
	if err != nil {
		log.Printf("Error cleaning up chirp: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   chirpClean.Body,
		UserID: chirpClean.UserID,
	}

	_, err = cfg.dbQueries.CreateChirp(context.Background(), chirpParams)
	if err != nil {
		log.Printf("Error saving chirp in database: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write(data)
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func chirpProfanityFilter(original Chirp) Chirp {
	profanities := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(original.Body, " ")
	for i, word := range words {
		_, ok := profanities[strings.ToLower(word)]
		if ok {
			words[i] = "****"
		} else {
			words[i] = word
		}
	}

	clean := strings.Join(words, " ")
	return Chirp{Body: clean, UserID: original.UserID}
}

func (cfg *apiConfig) userCreateHandler(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)

	type ReqValues struct {
		Email string `json:"email"`
	}
	reqValues := ReqValues{}
	err := decoder.Decode(&reqValues)
	if err != nil {
		log.Printf("Error decoding request: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := cfg.dbQueries.CreateUser(context.Background(), reqValues.Email)
	if err != nil {
		log.Printf("Error creating user: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	userValues := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UdpatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UdpatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	data, err := json.Marshal(userValues)
	if err != nil {
		log.Printf("Error after creating user: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write(data)
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(writer http.ResponseWriter, request *http.Request) {
	_ = request
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(writer, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) metricsResetHandler(writer http.ResponseWriter, request *http.Request) {
	_ = request
	if cfg.platform != "dev" {
		fmt.Println("dev plat")
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.dbQueries.DeleteAllUsers(context.Background())
	if err != nil {
		log.Printf("Error deleting users: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	cfg.fileserverHits.Store(0)
	writer.WriteHeader(http.StatusOK)
	fmt.Fprintln(writer, "Fileserver hits reset and users deleted")

}

func readyHandler(writer http.ResponseWriter, request *http.Request) {
	_ = request
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
