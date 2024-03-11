package database

import (
	"errors"
	"fmt"
	"time"
)

//type Token struct {
//	//ID   string    `json:"token_id"`
//	Time time.Time `json:"revoked_time"`
//}

var ErrTokenAlreadyRevoked = errors.New("already revoked")

func (db *DB) RevokeToken(token string) error {
	fmt.Println("Checking if Token exists in revoked list")
	revoked, err := db.GetTokenByToken(token)
	if err != nil {
		return err
	}
	fmt.Println("2")

	if revoked {
		return ErrTokenAlreadyRevoked
	}

	dbStructure, err := db.loadDB()
	fmt.Println("printing db before panic")
	fmt.Println(dbStructure)
	dbStructure.Tokens[token] = time.Now().UTC()
	fmt.Println("This only shows up if it did not panic")
	fmt.Println("3")

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	fmt.Println("4")

	return nil
}

func (db *DB) GetTokenByToken(token string) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}
	fmt.Println(dbStructure)
	_, exists := dbStructure.Tokens[token]
	if exists {
		fmt.Print("token is in the revoked list")
		return true, nil
	}

	fmt.Print("token was not found in the revoked list")

	return false, nil
}
