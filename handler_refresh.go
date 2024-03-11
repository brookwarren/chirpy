package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/brookwarren/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	subject, err := auth.ValidateJWT(token, cfg.jwtSecret, auth.RefreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	var revoked, _ = cfg.DB.GetTokenByToken(token)
	{
		if revoked == true {
			respondWithError(w, http.StatusUnauthorized, "Token revoked")
		}
	}

	subjectInt, _ := strconv.Atoi(subject)

	newToken, err := auth.MakeJWT(subjectInt, cfg.jwtSecret, time.Duration(1), auth.AccessToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	type response struct {
		TokenString string `json:"token"`
	}
	respondWithJSON(w, http.StatusOK, response{
		TokenString: newToken,
	})
}
