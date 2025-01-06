package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1) // increment the counter here
		next.ServeHTTP(w, r)      // call the next handler
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, http.StatusText(http.StatusOK))
}

func validate_length(w http.ResponseWriter, req *http.Request) {
	type paramsS struct {
		Body string `json:"body"`
	}

	type errorS struct {
		Error string `json:"error"`
	}

	type successS struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(req.Body)
	params := paramsS{}
	err := decoder.Decode(&params)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		errJson, _ := json.Marshal(errorS{Error: "Something went wrong"})

		log.Printf("Json error: %s", err)
		w.WriteHeader(500)
		w.Write(errJson)
	}

	if len(params.Body) <= 140 {
		successJson, _ := json.Marshal(successS{Valid: true})
		w.WriteHeader(200)
		w.Write(successJson)
	} else {
		errJson, _ := json.Marshal(errorS{Error: "Chirp is too long"})
		w.WriteHeader(400)
		w.Write(errJson)
	}
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	var apiCfg = apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/validate_chirp", validate_length)

	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.reset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
