package main

import (
   	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type apiHandler struct{}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func main() {
	mux := http.NewServeMux()
    corsMux := middlewareCors(mux)

	var srv http.Server

    srv.Addr = "192.168.69.175:8080"
    srv.Handler = corsMux

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed



	// mux.Handle("/api/", apiHandler{})
	// mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	// 	// The "/" pattern matches everything, so we need to check
	// 	// that we're at the root here.
	// 	if req.URL.Path != "/" {
	// 		http.NotFound(w, req)
	// 		return
	// 	}
	// 	fmt.Fprintf(w, "Welcome to the home page!")
	// })
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

