package main

import (
	"encoding/json"
	"fmt"
//	"io"
	"net/http"

	"strconv"
	"time"

	"github.com/brookwarren/chirpy/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

const DEFAULT_EXPIRATION_SECONDS = 24 * 60 * 60

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Password string `json:"password"`
        Email    string `json:"email"`
        Expires  int    `json:"expires_in_seconds"` // Changed type to int
    }

    type response struct {
        User
    }

    // Decode the request body into parameters struct
    decoder := json.NewDecoder(r.Body)
    defer r.Body.Close() // Close the request body after reading

    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
        return
    }

    // Use params.Expires directly instead of reading the request body again
    expiresInSeconds := params.Expires

    // Enforce maximum of 24 hours.
    if expiresInSeconds <= 0 || expiresInSeconds > DEFAULT_EXPIRATION_SECONDS {
        expiresInSeconds = DEFAULT_EXPIRATION_SECONDS
        fmt.Println("Expiration capped at 24 hours")
    }

    fmt.Println("expires in:", expiresInSeconds)

    // Get user from database
    user, err := cfg.DB.GetUserByEmail(params.Email)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't get user")
        return
    }

    // Check password
    err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid password")
        return
    }

    // Generate JWT token
    claims := &jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiresInSeconds) * time.Second)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        Issuer:    "chirpy",
        Subject:   strconv.Itoa(user.ID),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    ss, err := token.SignedString(cfg.JwtSecret)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Error generating JWT token")
        return
    }

    // Respond with token
    respondWithJSON(w, http.StatusOK, response{
        User: User{
            ID:    user.ID,
            Email: user.Email,
            Token: ss,
        },
    })
}

