package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/brookwarren/chirpy/internal/chirpydb"

	"github.com/go-chi/chi/v5"
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

	db, err := chirpydb.NewDB("database.json")
	if err != nil {
		fmt.Println("Error creating database:", err)
		return
	}

    fmt.Println("Main. Created DB: ", db)


	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
	apiRouter.Get("/chirps", handlerGetChirps(db))
	apiRouter.Post("/chirps", handlerCreateChirp(db))
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


func handlerCreateChirp(db *chirpydb.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    	type parameters struct {
	    	Body string `json:"body"`
    	}

	    decoder := json.NewDecoder(r.Body)
    	params := parameters{}
    	err := decoder.Decode(&params)
    	if err != nil {
	    	respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		    return
    	}

    	// Call CreateChirp to create a new chirp
	    chirp, err := db.CreateChirp(params.Body)
    	if err != nil {
		    respondWithError(w, http.StatusBadRequest, err.Error())
	    	return
    	}

	    // Respond with the created chirp
    	respondWithJSON(w, http.StatusCreated, chirp)
    }
}


func handlerGetChirps(db *chirpydb.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Retrieve all chirps from the database
        chirps, err := db.GetChirps()
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "Failed to get chirps from database")
            return
        }

        // Create slice to hold chirp response objects
        var response []map[string]interface{}

        // Iterate over chirps and construct response objects
        for _, chirp := range chirps {
            chirpResponse := map[string]interface{}{
                "id":   chirp.ID,
                "body": chirp.Body,
            }
            response = append(response, chirpResponse)
        }

        // // Format chirps as JSON
        // type chirpResponse struct {
        //     ID   int    `json:"id"`
        //     Body string `json:"body"`
        // }

        // var response []chirpResponse
        // for _, chirp := range chirps {
        //     response = append(response, chirpResponse{
        //         ID:   chirp.ID,
        //         Body: chirp.Body,
        //     })
        // }

        // Respond with JSON
        respondWithJSON(w, http.StatusOK, response)
    }
}





func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
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

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(params.Body, badWords)

	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: cleaned,
	})
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
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
