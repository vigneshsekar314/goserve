package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/vigneshsekar314/goserve/internal/auth"
	"github.com/vigneshsekar314/goserve/internal/database"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
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
func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirp_id := r.PathValue("chirp_id")
	if chirp_id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid API path"))
		return
	}
	uuidChirp, err := uuid.Parse(chirp_id)
	if err != nil {
		log.Printf("unable to parse uuid from url: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chirpData, err := cfg.dbconfig.GetChirp(r.Context(), uuidChirp)

	if err != nil {
		log.Printf("chirp not found: %s\n", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	chirpRes := ChirpResponse{
		Id:        chirpData.ID,
		Body:      chirpData.Body,
		CreatedAt: chirpData.CreatedAt,
		UpdatedAt: chirpData.UpdatedAt,
		UserId:    chirpData.UserID,
	}
	data, err := json.Marshal(chirpRes)
	if err != nil {
		log.Printf("error marshaling chirpData: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbconfig.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("error retrieving chirps: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	allChirps := make([]ChirpResponse, 0)
	for _, chirp := range chirps {
		allChirps = append(allChirps, ChirpResponse{
			Id:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
		})
	}
	data, err := json.Marshal(allChirps)
	if err != nil {
		log.Printf("error parsing all chirps: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq createAndLoginUserReq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginReq); err != nil {
		log.Printf("error in decoding user request %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Request is invalid"))
		return
	}
	user, err := cfg.dbconfig.GetUser(r.Context(), loginReq.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(EMAILPASSERR))
		return
	}
	match, err := auth.CheckPasswordHash(loginReq.Password, user.HashedPassword)
	if err != nil || !match {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(EMAILPASSERR))
		return
	}
	loginResp := createLoginUserRes{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respBytes, err := json.Marshal(loginResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling response %s\n", err)
		return
	}
	w.Write(respBytes)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {

	var createRq createAndLoginUserReq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&createRq); err != nil {
		w.WriteHeader(500)
		fmt.Printf("error parsing request, %s", err)
		return
	}
	// hash password
	hashedPwd, err := auth.HashPassword(createRq.Password)
	if err != nil {
		log.Printf("error in hashing password: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, err := cfg.dbconfig.CreateUser(r.Context(), database.CreateUserParams{
		Email:          createRq.Email,
		HashedPassword: hashedPwd,
	})
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("error creating user, %s", err)
		return
	}
	createUserRsp := createLoginUserRes{
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
