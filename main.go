package main

import (
	"log"
	"net/http"
	"sync"
    "fmt"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
    // apiConfigWithMetrics := apiConfig{}
    apiCfg := apiConfig{}
    // apiCfg := apiConfigWithMetrics.middlewareMetricsInc(mux)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("/healthz", handlerReadiness)
	mux.HandleFunc("/metrics", apiCfg.hitCount)
	mux.HandleFunc("/reset", apiCfg.resetCount)

	corsMux := middlewareCors(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}


type apiConfig struct {
	fileserverHits int
    mu  sync.Mutex
}

func (cfg *apiConfig) hitCount(w http.ResponseWriter, r *http.Request) {
    cfg.mu.Lock()
    defer cfg.mu.Unlock()

    fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits)
}

func (cfg *apiConfig) resetCount(w http.ResponseWriter, r *http.Request) {
    cfg.mu.Lock()
    cfg.fileserverHits = 0
    cfg.mu.Unlock()
    fmt.Fprintf(w, "Hits reset to: %d", cfg.fileserverHits)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.mu.Lock()
        cfg.fileserverHits++
        cfg.mu.Unlock()
        next.ServeHTTP(w, r)
    })
}
