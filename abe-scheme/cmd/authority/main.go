package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/crypto"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils/policyConfig"
)

var scheme *crypto.ABEscheme

const databaseURL = "http://localhost:8080"
const authorityUUID = "497dcba3-ecbf-4587-a2dd-5eb0665e6880"

func main() {

	scheme = crypto.Setup()
	updatePolicyConfig()

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

func updatePolicyConfig() {

	newPolicyConfig := policyConfig.Config{
		PurposeTrees: utils.ExamplePurposeTrees(),
		Scheme:       crypto.ABEscheme{PublicKey: scheme.PublicKey},
	}

	newRecord := utils.Record{
		Table:           "relations",
		ID:              utils.Assure(uuid.Parse(authorityUUID)),
		PrivateWriteKey: []byte{},
		PublicWriteKey:  []byte{},
		Data:            utils.ToBytes(newPolicyConfig),
		Created:         time.Now(),
	}

	jsonData := utils.Assure(json.Marshal(newRecord))
	resp := utils.Assure(http.Post(databaseURL+"/entries", "application/json", bytes.NewBuffer(jsonData)))
	defer resp.Body.Close()

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("entry add failed: %s", body)
	}
}
