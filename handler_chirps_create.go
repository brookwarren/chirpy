package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/brookwarren/chirpy/internal/auth"
)

type Chirp struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	Author_ID int    `json:"author_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

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

	author_id, validate_err := auth.ValidateJWT(token, cfg.jwtSecret)
	if validate_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	authorIdInt, _ := strconv.Atoi(author_id)

	// fmt.Println(author_id)

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.DB.CreateChirp(cleaned, authorIdInt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		Author_ID: authorIdInt,
	})
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
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
