package main

import (
	"net/http"
	"sort"
	"strconv"
	"github.com/brookwarren/chirpy/internal/database"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:   dbChirp.ID,
			Body: dbChirp.Body,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpsRetrieveByID(w http.ResponseWriter, r *http.Request) {
    // Extract chirpid from URL
    chirpID := chi.URLParam(r, "chirpid")


    // Convert chirpid to integer
    id, err := strconv.Atoi(chirpID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
    }

    // Retrieve chirp from database
    chirp, err := cfg.DB.GetChirpByID(id)
    if err != nil {
        if err == database.ErrChirpNotFound {
            respondWithError(w, http.StatusNotFound, "Chirp not found")
            return
        }
        respondWithJSON(w, http.StatusOK, chirp)
        return
    }

   	// Respond with chirp
	respondWithJSON(w, http.StatusOK, chirp)
}

