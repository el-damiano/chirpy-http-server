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
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *apiConfig) chirpCreateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding chirp: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirpValidated := len(chirp.Body) <= 140
	if !chirpValidated {
		w.WriteHeader(http.StatusBadRequest)
	}

	chirpClean := chirpProfanityFilter(chirp)
	chirpParams := database.CreateChirpParams{
		Body:   chirpClean.Body,
		UserID: chirpClean.UserID,
	}

	chirpDB, err := cfg.dbQueries.CreateChirp(context.Background(), chirpParams)
	if err != nil {
		log.Printf("Error saving chirp in database: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirpValues := Chirp{
		ID:        chirpDB.ID,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
	}

	data, err := json.Marshal(chirpValues)
	if err != nil {
		log.Printf("Error cleaning up chirp: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	if id != "" {
		cfg.chirpWriteByID(w, r, id)
		return
	}

	chirps, err := cfg.dbQueries.GetAllChirps(context.Background())
	if err != nil {
		log.Printf("Error retrieving all the chirps: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var chirpsResponse []Chirp
	for _, chirp := range chirps {
		chirpy := Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}
		chirpsResponse = append(chirpsResponse, chirpy)
	}

	data, err := json.Marshal(chirpsResponse)
	if err != nil {
		log.Printf("Error encoding the retrieved chirps: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) chirpWriteByID(w http.ResponseWriter, r *http.Request, id string) {
	_ = r

	uuid, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Error retrieving the chirp of ID %s: %s\n", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	chirpDB, err := cfg.dbQueries.GetChirpByID(context.Background(), uuid)
	if err != nil {
		log.Printf("Error retrieving the chirp of ID %s: %s\n", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirp := Chirp{
		ID:        chirpDB.ID,
		Body:      chirpDB.Body,
		UpdatedAt: chirpDB.UpdatedAt,
		CreatedAt: chirpDB.CreatedAt,
		UserID:    chirpDB.UserID,
	}
	data, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error encoding the retrieved chirp: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	return Chirp{
		Body:      clean,
		UserID:    original.UserID,
		ID:        original.ID,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}
}

func (cfg *apiConfig) userCreateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	type ReqValues struct {
		Email string `json:"email"`
	}
	reqValues := ReqValues{}
	err := decoder.Decode(&reqValues)
	if err != nil {
		log.Printf("Error decoding request: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := cfg.dbQueries.CreateUser(context.Background(), reqValues.Email)
	if err != nil {
		log.Printf("Error creating user: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) metricsResetHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	if cfg.platform != "dev" {
		fmt.Println("dev plat")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.dbQueries.DeleteAllUsers(context.Background())
	if err != nil {
		log.Printf("Error deleting users: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Fileserver hits reset and users deleted")

}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
