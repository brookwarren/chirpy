package main

import (
	"fmt"
	"github.com/brookwarren/chirpy/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handlerTokenRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	_, validate_err := auth.ValidateJWT(token, cfg.jwtSecret, auth.RefreshToken)
	if validate_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	//fmt.Println("attempting to revoke token****************************************")
	revoke_err := cfg.DB.RevokeToken(token)
	if revoke_err != nil {
		return
	}
	fmt.Println("succeeded at revoking token***************************************")

	respondWithJSON(w, http.StatusOK, "")
}
