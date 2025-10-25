package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/vigneshsekar314/goserve/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbconfig       *database.Queries
	environment    string
	auth_token     string
}

func main() {
	godotenv.Load()
	serveMux := http.NewServeMux()
	dbURL := os.Getenv("DB_URL")
	btoken := os.Getenv("BTOKEN")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error connect to db: %s", err)
	}
	dbQueries := database.New(db)
	cf := apiConfig{
		dbconfig:    dbQueries,
		environment: os.Getenv("PLATFORM"),
		auth_token:  btoken,
	}
	serveMux.Handle("/app/", cf.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", healthStatus)
	serveMux.HandleFunc("GET /admin/metrics", cf.readServerHits)
	serveMux.HandleFunc("POST /admin/reset", cf.resetServerHits)
	serveMux.HandleFunc("POST /api/users", cf.handleUsers)
	serveMux.HandleFunc("POST /api/login", cf.handleLogin)
	serveMux.HandleFunc("GET /api/chirps/{chirp_id}", cf.handleGetChirp)
	serveMux.HandleFunc("GET /api/chirps", cf.handleGetAllChirps)
	serveMux.HandleFunc("POST /api/chirps", cf.handleCreateChirp)
	httpServe := http.Server{Handler: serveMux, Addr: ":8080"}
	log.Printf("Server listening on port %v", httpServe.Addr)
	log.Fatal(httpServe.ListenAndServe())
}
