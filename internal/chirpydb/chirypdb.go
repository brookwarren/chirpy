package chirpydb

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDB()
	if err != nil {
		return nil, err
	}
    // fmt.Println("Exiting newDB")
	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	// Validate chirp body length
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return Chirp{}, fmt.Errorf("chirp is too long")
	}

	// Replace bad words in chirp body
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleanedBody := strings.Join(words, " ")


	dbData, err := db.loadDB()
	if err != nil {
		return Chirp{}, fmt.Errorf("error reading existing chirps: %v", err)
	}

	nextID := len(dbData.Chirps) + 1

	newChirp := Chirp{
		ID:   nextID,
		Body: cleanedBody,
	}

	dbData.Chirps[nextID] = newChirp

	err = db.writeDB(dbData)
	if err != nil {
		return Chirp{}, fmt.Errorf("error writing updated chirps: %v", err)
	}

	return newChirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	// db.mux.RLock()
	// defer db.mux.RUnlock()

	dbData, err := db.loadDB()
	if err != nil {
		return nil, fmt.Errorf("error reading existing chirps: %v", err)
	}

	chirps := make([]Chirp, 0, len(dbData.Chirps))
	for _, chirp := range dbData.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		// Create an empty database file
		err := db.writeDB(DBStructure{})
		if err != nil {
			return fmt.Errorf("error creating and writing to database file: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking database file: %v", err)
	}
	return nil
}



// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	// Read from the database file
	data, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, fmt.Errorf("error reading database file: %v", err)
	}

	// If the database is empty, return an empty DBStructure
	if len(data) == 0 {
		return DBStructure{Chirps: make(map[int]Chirp)}, nil
	}

	// Unmarshal the data into DBStructure
	var dbData DBStructure
	err = json.Unmarshal(data, &dbData)
	if err != nil {
		return DBStructure{}, fmt.Errorf("error unmarshalling database data: %v", err)
	}

	// Ensure Chirps map is initialized
	if dbData.Chirps == nil {
		dbData.Chirps = make(map[int]Chirp)
	}

	return dbData, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling database data: %v", err)
	}

	err = os.WriteFile(db.path, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing database file: %v", err)
	}

	return nil
}
