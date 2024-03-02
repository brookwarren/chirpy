package main

import (
	"log"
	"net/http"
	"github.com/go-chi/chi/v5"
    "encoding/json"
    "strings"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
	apiRouter.Post("/validate_chirp", handlerChirpsValidate)
	router.Mount("/api", apiRouter)

    adminRouter := chi.NewRouter()
    adminRouter.Get("/metrics", apiCfg.handlerMetrics)
    router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}


func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type cleaned struct {
		Cleaned_Body string `json:"cleaned_body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

    if filterAndReplace(&params.Body) {
	    respondWithJSON(w, http.StatusOK, cleaned{
		    Cleaned_Body: params.Body,
    	})
    } else {
    	respondWithJSON(w, http.StatusOK, returnVals{
	    	Valid: true,
    	})
    }
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func filterAndReplace(input *string) bool {

    badWords := []string{"kerfuffle", "sharbert", "fornax"}
    profane := false
    words := strings.Split(*input, " ") 
    for i, word := range words {
        for _, badWord := range badWords {
            if strings.ToLower(word) == badWord {
                words[i] = "****"
                profane = true
            }
        }
    }
    if profane {
        *input = strings.Join(words, " ")
    }
    return profane
}
