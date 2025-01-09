package main

import (
	"chirpy/internal/database"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET is required")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM is required")
	}

	dbconn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatalf("Could not reach database: %s", err)
	}

	dbQueries := database.New(dbconn)

	var apiCfg = apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, http.StatusText(http.StatusOK))
	})
	mux.HandleFunc("GET /api/chirps", apiCfg.get_chirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.get_chirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.create_chirp)
	mux.HandleFunc("POST /api/users", apiCfg.create_user)
	mux.HandleFunc("POST /api/login", apiCfg.login_user)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
