package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/el-damiano/bootdev-http-server/internal/auth"
	"github.com/el-damiano/bootdev-http-server/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	platform       string
	tokenSecret    string
	polkaKey       string
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
	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding chirp failed", err)
		return
	}

	tokenBearer, err := auth.BearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Authorization failed: %v", err), err)
		return
	}

	userID, err := auth.ValidateJWT(tokenBearer, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization failed: invalid/expired JWT", err)
		return
	}

	chirpClean, err := chirpValidate(chirp.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
	}

	chirpParams := database.CreateChirpParams{
		Body:   chirpClean,
		UserID: userID,
	}

	chirpDB, err := cfg.dbQueries.CreateChirp(context.Background(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Saving chirp failed", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirpDB.ID,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
	})
}

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	if id != "" {
		cfg.chirpWriteByID(w, r, id)
		return
	}

	var chirps []database.Chirp
	var err error

	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString != "" {
		var authorID uuid.UUID
		authorID, err = uuid.Parse(authorIDString)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Invalid author_id", err)
			return
		}
		chirps, err = cfg.dbQueries.GetUsersChirps(context.Background(), authorID)
	} else {
		chirps, err = cfg.dbQueries.GetAllChirps(context.Background())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving all the chirps", err)
		return
	}

	sortOrder := r.URL.Query().Get("sort") // by default chirps are sorted by asc
	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
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

	respondWithJSON(w, http.StatusOK, chirpsResponse)
}

func (cfg *apiConfig) chirpsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	tokenBearer, err := auth.BearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization failed: invalid/expired JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(tokenBearer, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization failed: invalid/expired JWT", err)
		return
	}

	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirpDB, err := cfg.dbQueries.GetChirpByID(context.Background(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Getting chirp failed", err)
		return
	}
	if userID != chirpDB.UserID {
		respondWithError(w, http.StatusForbidden, "Authorization error: insufficient permissions to delete chirp", err)
		return
	}

	err = cfg.dbQueries.DeleteChirpByID(context.Background(), chirpDB.ID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Deleting chirp failed", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) chirpWriteByID(w http.ResponseWriter, r *http.Request, id string) {
	_ = r

	uuid, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Getting chirp failed", err)
		return
	}
	chirpDB, err := cfg.dbQueries.GetChirpByID(context.Background(), uuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Getting chirp failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirpDB.ID,
		Body:      chirpDB.Body,
		UpdatedAt: chirpDB.UpdatedAt,
		CreatedAt: chirpDB.CreatedAt,
		UserID:    chirpDB.UserID,
	})
}

func chirpValidate(bodyOriginal string) (string, error) {
	const chirpLenMax = 140
	if len(bodyOriginal) > chirpLenMax {
		return "", errors.New("Chirp is too long")
	}

	profanities := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(bodyOriginal, " ")
	for i, word := range words {
		_, ok := profanities[strings.ToLower(word)]
		if ok {
			words[i] = "****"
		} else {
			words[i] = word
		}
	}

	bodyClean := strings.Join(words, " ")
	return bodyClean, nil
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UdpatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Token        string    `json:"token"`
	TokenRefresh string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) userCreateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	type ReqValues struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	reqValues := ReqValues{}
	err := decoder.Decode(&reqValues)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request", err)
		return
	}

	if reqValues.Email == "" || reqValues.Password == "" {
		respondWithError(
			w,
			http.StatusUnprocessableEntity,
			"'email' and 'password' are required",
			errors.New("User did not provide emmail and password"))
		return
	}

	passwordHashed, err := auth.HashPassword(reqValues.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	params := database.CreateUserParams{
		Email:          reqValues.Email,
		HashedPassword: passwordHashed,
	}

	user, err := cfg.dbQueries.CreateUser(context.Background(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UdpatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	tokenBearer, err := auth.BearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Authorization failed: %v", err), err)
		return
	}
	userID, err := auth.ValidateJWT(tokenBearer, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization failed: invalid/expired JWT", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	type UpdateRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	updateRequest := UpdateRequest{}
	err = decoder.Decode(&updateRequest)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request", err)
		return
	}

	passwordHashed, err := auth.HashPassword(updateRequest.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	userUpdateParams := database.UpdateUserParams{
		Email:          updateRequest.Email,
		HashedPassword: passwordHashed,
		ID:             userID,
	}
	userDB, err := cfg.dbQueries.UpdateUser(context.Background(), userUpdateParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:          userDB.ID,
		CreatedAt:   userDB.CreatedAt,
		UdpatedAt:   userDB.UpdatedAt,
		Email:       userDB.Email,
		IsChirpyRed: userDB.IsChirpyRed,
	})
}

func (cfg *apiConfig) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	type ReqValues struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	reqValues := ReqValues{}
	err := decoder.Decode(&reqValues)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request", err)
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(context.Background(), reqValues.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(reqValues.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	tokenJWT, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT token", err)
		return
	}

	tokenRefresh, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making refresh token", err)
		return
	}

	tokenRefreshParams := database.CreateTokenParams{
		Token:     tokenRefresh,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}

	_, err = cfg.dbQueries.CreateToken(context.Background(), tokenRefreshParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UdpatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        tokenJWT,
		TokenRefresh: tokenRefresh,
	})
}

func (cfg *apiConfig) tokenRefreshHandler(w http.ResponseWriter, r *http.Request) {
	tokenBearer, err := auth.BearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Authorization failed: %v", err), err)
		return
	}

	userId, err := cfg.dbQueries.GetUserFromRefreshToken(context.Background(), tokenBearer)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization failed, token doesn't exist or is expired", err)
		return
	}

	tokenJWT, err := auth.MakeJWT(userId, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: tokenJWT,
	})
}

func (cfg *apiConfig) tokenRevokeHandler(w http.ResponseWriter, r *http.Request) {
	tokenBearer, err := auth.BearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Authorization failed: %v", err), err)
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(context.Background(), tokenBearer)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Token revokation failed", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) polkaHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.APIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Authorization failed: %v", err), err)
		return
	}

	if apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.dbQueries.UpgradeUserMembership(context.Background(), params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	_, err := fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	if err != nil {
		log.Printf("Error writing message: %s", err)
		return
	}
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
	_, err = fmt.Fprintln(w, "Fileserver hits reset and users deleted")
	if err != nil {
		log.Printf("Error writing message: %s", err)
		return
	}
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("Error writing to the HTTP response: %s\n", err)
		return
	}
}
