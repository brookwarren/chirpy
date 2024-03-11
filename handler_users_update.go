package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/brookwarren/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	subject, err := auth.ValidateJWT(token, cfg.jwtSecret, auth.AccessToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	//fmt.Println("1")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	//fmt.Println("2")

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	//fmt.Println("3")

	userIDInt, err := strconv.Atoi(subject)
	if err != nil {
		//respondWithError(w, http.StatusInternalServerError, "Couldn't parse user ID")
		respondWithError(w, http.StatusUnauthorized, "Couldn't parse user ID")
		return
	}

	//fmt.Println("4")

	user, err := cfg.DB.UpdateUser(userIDInt, params.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	//fmt.Println("5")

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}
