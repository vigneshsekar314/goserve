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
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()
	serveMux := http.NewServeMux()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error connect to db: %s", err)
	}
	dbQueries := database.New(db)
	cf := apiConfig{
		dbconfig: dbQueries,
	}
	serveMux.Handle("/app/", cf.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", (func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	serveMux.HandleFunc("GET /admin/metrics", cf.readServerHits)
	serveMux.HandleFunc("POST /admin/reset", cf.resetServerHits)
	serveMux.HandleFunc("POST /api/validate_chirp", validate_chirp)
	httpServe := http.Server{Handler: serveMux, Addr: ":8080"}
	log.Fatal(httpServe.ListenAndServe())
}
