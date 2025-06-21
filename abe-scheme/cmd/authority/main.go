package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/crypto"
)

type Record struct {
	Table           string    `json:"table"`
	ID              string    `json:"id"`
	PrivateWriteKey []byte    `json:"private_write_key"`
	PublicWriteKey  []byte    `json:"public_write_key"`
	Data            []byte    `json:"data"`
	Created         time.Time `json:"created"`
}

var scheme *crypto.ABEscheme

func main() {

	scheme = crypto.Setup()

	r := mux.NewRouter()
	r.HandleFunc("/get_key", getKey).Methods("GET")

	log.Println("key authority server started on port :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func getKey(w http.ResponseWriter, r *http.Request) {
	attributes := r.URL.Query()["attribute"]
	fmt.Printf("generating key for attributes %v\n", attributes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scheme.KeyGen(attributes))
}
