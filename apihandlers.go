package main

import (
	"encoding/json"
	"fmt"
	"github.com/vigneshsekar314/goserve/internal/database"
	"log"
	"net/http"
	"os"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleChirps(w http.ResponseWriter, r *http.Request) {
	var chirp ChirpRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirp); err != nil {
		w.WriteHeader(400)
		log.Printf("error in decoding request: %s/n", err)
		return
	}
	log.Printf(".user_id: %s and .body: %s", chirp.UserId, chirp.Body)

	validJson, err := validate_chirp(chirp)
	if err != nil {
		w.WriteHeader(400)
		log.Printf("error in validation: %s/n", err)
		res, err := json.Marshal(ErrorJson{Error: err.Error()})
		if err != nil {
			log.Printf("error in error marshal: %s/n", err)
			w.WriteHeader(400)
			return
		}
		w.Write(res)
		return
	}

	newChirp, err := cfg.dbconfig.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validJson.CleanedBody,
		UserID: chirp.UserId,
	})
	chirpResponse := ChirpResponse{
		Id:        newChirp.ID,
		Body:      newChirp.Body,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		UserId:    newChirp.UserID,
	}
	if err != nil {
		w.WriteHeader(500)
		log.Printf("error creating chirp: %s\n", err)
		return
	}
	newChirpBytes, err := json.Marshal(chirpResponse)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("error in marshaling newChirp: %s/n", err)
		return
	}
	w.WriteHeader(201)
	w.Write(newChirpBytes)
}

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {

	var createRq createUserReq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createRq); err != nil {
		w.WriteHeader(500)
		fmt.Printf("error parsing request, %s", err)
		return
	}
	user, err := cfg.dbconfig.CreateUser(r.Context(), createRq.Email)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("error creating user, %s", err)
		return
	}
	createUserRsp := createUserRes{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	rs, err := json.Marshal(createUserRsp)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("error marshalling created user, %s", err)
		return
	}
	w.WriteHeader(201)
	w.Write(rs)
}

func (cfg *apiConfig) readServerHits(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("./metrics.html")
	if err != nil {
		fmt.Fprintf(w, "page not found")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, fmt.Sprintf(string(content), cfg.fileserverHits.Load()))
}
func (cfg *apiConfig) resetServerHits(w http.ResponseWriter, r *http.Request) {
	// cfg.fileserverHits.Store(0)
	if cfg.environment != "dev" {
		w.WriteHeader(403)
		return
	}
	if err := cfg.dbconfig.DeleteUsers(r.Context()); err != nil {
		w.WriteHeader(500)
		fmt.Printf("error occured when deleting all users: %s\n", err)
		return
	}
	w.Write([]byte("Reset done"))
}
func healthStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
