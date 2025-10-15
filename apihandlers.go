package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
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
