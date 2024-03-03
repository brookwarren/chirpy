package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"fmt"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
    "sort"
)

type apiConfig struct {
	fileserverHits int
    id int
}

type Chirp struct {
    id int
    body string
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}


func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
        id: 0,
	}

    db, err := NewDB("database.json")
    if err != nil {
		fmt.Println("Error creating database:", err)
		os.Exit(1) 
        return
    }

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
    apiRouter.Post("/chirps", db.handlerChirpsValidate)
	apiRouter.Get("/chirps", db.chirpsHandler)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)
	router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (db *DB) handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxChirpLength = 140
	if len(chirp.body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}


    chirp.body = getCleanedBody(chirp.body, badWords)

    newChirp, err := db.CreateChirp("chirp.body")

    respondWithJSON(w, http.StatusCreated, Chirp{
        body: newChirp.body,
        id: newChirp.id,
	})
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}



// DATABASE HANDLING. THIS WILL GO INTO A PACKAGE LATER




// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	if err := db.ensureDB(); err != nil {
		return nil, err
	}
	return &db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbData, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	// Get next ID (could be optimized, but works with current setup)
	nextID := 1
	for id := range dbData.Chirps {
		if id >= nextID {
			nextID = id + 1
		}
	}

	newChirp := Chirp{id: nextID, body: body}
	dbData.Chirps[nextID] = newChirp

	err = db.writeDB(dbData)
	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

// GetChirps returns all chirps from the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbData, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbData.Chirps))
	for _, chirp := range dbData.Chirps {
		chirps = append(chirps, chirp)
	}

	// Optional: Sort chirps by ID
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].id < chirps[j].id
	})

	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		// Create an empty DBStructure
		dbData := DBStructure{Chirps: make(map[int]Chirp)}

		if err := db.writeDB(dbData); err != nil {
			return err
		}
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	data, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

    fmt.Println(data)

	var dbData DBStructure
	err = json.Unmarshal(data, &dbData)
	if err != nil {
		return DBStructure{}, err
	}

	return dbData, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbData DBStructure) error {
	data, err := json.Marshal(dbData)
	if err != nil {
		return err
	}
	return os.WriteFile(db.path, data, 0644) 
}

func (db *DB) chirpsHandler(w http.ResponseWriter, r *http.Request) {
    chirps, err := db.GetChirps()
    if err != nil {
        http.Error(w, "Error fetching chirps: "+err.Error(), http.StatusInternalServerError)
        return
    }

    chirpsJSON, err := json.Marshal(chirps)
    if err != nil {
        http.Error(w, "Error encoding chirps to JSON: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(chirpsJSON)
}





// // NewDB creates a new database connection
// // and creates the database file if it doesn't exist
// func NewDB(path string) (*DB, error) {
//     db := DB{
//         path: path,
//         mux: &sync.RWMutex{},
//     }

//     _, err := os.Stat(path)
//     if os.IsNotExist(err) {
//         _, err := os.Create(path)
//         if err != nil {
//             return &db, err
//         }
//     }
//     return &db, nil
// }

// // CreateChirp creates a new chirp and saves it to disk
// func (db *DB) CreateChirp(body string) (Chirp, error) {

//     chirps, err := db.GetChirps()
//     if err != nil {
//         return Chirp{}, nil
//     }

//     nextChirpId := len(chirps) + 1
//     newChirp := Chirp{
//         id: nextChirpId,
//         body: body,
//     }

//     chirps = append(chirps, newChirp)
//     data, err := json.Marshal(chirps)
//     if err != nil {
//         return Chirp{}, err
//     }

//     err = os.WriteFile(db.path, data, 0644)
//     if err != nil {
//         return Chirp{}, err
//     }


//     // chirp := Chrip {
//     //     Body: body,
//     //     Id: id,
//     // }
//      return Chirp{}, nil
// }

// // GetChirps returns all chirps in the database
// func (db *DB) GetChirps() ([]Chirp, error) {
//     db.mux.RLock()
//     defer db.mux.RUnlock()

//     data, err := os.ReadFile(db.path)
//     if err != nil {
//         if os.IsNotExist(err) {
//             return []Chirp{}, nil
//         }
//         return nil, err
//     }

//     var dbData DBStructure
//     err = json.Unmarshal(data, &dbData)
//     if err != nil {
//         return nil, err
//     }

//     return chirps, nil
// }

// // // ensureDB creates a new database file if it doesn't exist
// // func (db *DB) ensureDB() error {
// //     return
// // }

// // // loadDB reads the database file into memory
// // func (db *DB) loadDB() (DBStructure, error) {
// //     return
// // }

// // // writeDB writes the database file to disk
// // func (db *DB) writeDB(dbStructure DBStructure) error {
// //     return
// // }
