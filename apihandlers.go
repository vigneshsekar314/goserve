package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error fetching token: %s\n", err)
		http.Error(w, "error fetching auth token", http.StatusUnauthorized)
		return
	}
	user_id, err := auth.ValidateJWT(token, cfg.auth_token)
	if err != nil {
		log.Printf("user unauthorized: %s\n", err)
		http.Error(w, "auth token unauthorized", http.StatusUnauthorized)
		return
	}

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
		UserID: user_id,
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

// handleLogin authenticates a user by validating their email and password.
// It expects a JSON request body with email and password fields.
// Returns 200 with user details on successful authentication,
// 400 for invalid request format, or 401 for invalid credentials.
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
	// generate jwt
	signedToken, err := auth.MakeJWT(user.ID, cfg.auth_token, time.Duration(time.Second*60*60))
	if err != nil {
		log.Printf("unable to generate JWT: %s\n", err)
		http.Error(w, "unable to generate JWT", http.StatusInternalServerError)
		return
	}
	// generate refresh token
	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("unable to generate refresh token: %s\n", err)
		http.Error(w, "unable to generate refresh token", http.StatusInternalServerError)
		return
	}
	_, err = cfg.dbconfig.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		log.Printf("unable to save refresh token: %s\n", err)
		http.Error(w, "unable to save refresh token", http.StatusInternalServerError)
		return
	}
	loginResp := createLoginUserRes{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        signedToken,
		RefreshToken: refresh_token,
	}
	respBytes, err := json.Marshal(loginResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling response %s\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("invalid / no refresh token found: %s\n", err)
		http.Error(w, "invalid / no refresh token found in headers", http.StatusUnauthorized)
		return
	}
	user_id_from_db, err := cfg.dbconfig.GetAccessTokenFromRefreshToken(r.Context(), refresh_token)
	if err != nil {
		log.Printf("refresh token expired or not available in database%s\n", err)
		http.Error(w, "refresh token expired or not available in database", http.StatusUnauthorized)
		return
	}
	access_token, err := auth.MakeJWT(user_id_from_db, cfg.auth_token, time.Duration(time.Second*60*60))
	if err != nil {
		log.Printf("unable to generate access token: %s\n", err)
		http.Error(w, "unable to generate access token", http.StatusUnauthorized)
		return
	}
	resp_bytes, err := json.Marshal(auth.RefreshTokenResponse{
		Token: access_token,
	})
	if err != nil {
		log.Printf("unable to marshall access_token: %s\n", err)
		http.Error(w, "unable to marshall access_token", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp_bytes)
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("invalid / no refresh token found: %s\n", err)
		http.Error(w, "invalid / no refresh token found in headers", http.StatusUnauthorized)
		return
	}
	_, err = cfg.dbconfig.RevokeRefreshToken(r.Context(), refresh_token)
	if err != nil {
		log.Printf("error on revoking refresh token: %s\n", err)
		http.Error(w, "error on revoking refresh token", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
