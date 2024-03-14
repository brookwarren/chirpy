package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/brookwarren/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	token, token_err := auth.GetBearerToken(r.Header)
	if token_err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find JWT")
		return
	}

	isRevoked, revoked_err := cfg.DB.IsTokenRevoked(token)
	if revoked_err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't check session")
		return
	}
	if isRevoked {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is revoked")
		return
	}

	author_id_str, validate_err := auth.ValidateJWT(token, cfg.jwtSecret)
	if validate_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	author_id, _ := strconv.Atoi(author_id_str)

	chirpID_str := r.PathValue("chirpID")
	chirpID, chirp_id_err := strconv.Atoi(chirpID_str)
	if chirp_id_err != nil {
		fmt.Print("ChirpIP is jacked up")
	}

	chirp, get_err := cfg.DB.GetChirp(chirpID)
	if get_err != nil {
		return
	}

	if chirp.Author_ID != author_id {
		respondWithJSON(w, http.StatusForbidden, Chirp{})
	}

	cfg.DB.DeleteChirp(chirpID)

	respondWithJSON(w, http.StatusOK, Chirp{})
}
