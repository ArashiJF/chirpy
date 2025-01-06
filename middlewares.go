package main

import "net/http"

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1) // increment the counter here
		next.ServeHTTP(w, r)      // call the next handler
	})
}
