package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/brookwarren/chirpy/internal/auth"
	"github.com/brookwarren/chirpy/internal/database"
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
    Token    string `json:"token"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
        Token    string `json:"token"`
	}
	// type response struct {
	// 	User
	// }

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters 2")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	user, err := cfg.DB.CreateUser(params.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, database.ErrAlreadyExists) {
			respondWithError(w, http.StatusConflict, "User already exists")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

    type response struct {
        ID      int    `json:"id"`
        Email   string `json:"email"`
    }

	respondWithJSON(w, http.StatusCreated, response{
			Email: user.Email,
			ID:    user.ID,
		})
}


func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Missing Authorization header")
		return
	}

	tokenString := authHeader[len("Bearer "):] // Assumes the header format is "Bearer <token>"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what you expect (HMAC SHA256 in your case)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return cfg.JwtSecret, nil
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token in handler_users_create/handlerUsersUpdate")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid token claims")
		return
	}

	userID := claims["sub"].(string)


	type parameters struct {
        ID       int    `json:"id"`
        Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	decode_err := decoder.Decode(&params)
	if decode_err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters 1")
		return
    }

    hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

    id, _ := strconv.Atoi(userID)

    userUpdate, err := cfg.DB.UpdateUser(id, params.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, database.ErrAlreadyExists) {
			respondWithError(w, http.StatusConflict, "User already exists")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

    type response struct {
        ID      int    `json:"id"`
        Email   string `json:"email"`
    }

	respondWithJSON(w, http.StatusOK, response{
			ID:    userUpdate.ID,
			Email: userUpdate.Email,
    })
}
