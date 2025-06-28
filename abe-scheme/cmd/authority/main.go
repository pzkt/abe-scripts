/*

The authority is responsible for setting up the ABE scheme and generating private keys
They must also update the policy config

*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/crypto"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils/policyConfig"
)

var scheme *crypto.ABEscheme
var setup_time int64

const databaseURL = "http://localhost:8080"
const authorityUUID = "497dcba3-ecbf-4587-a2dd-5eb0665e6880"

func main() {
	setup_time = time.Now().Unix()

	scheme = crypto.Setup()
	updatePolicyConfig()

	r := mux.NewRouter()
	r.HandleFunc("/get_key", getKey).Methods("GET")
	r.HandleFunc("/get_time_key", getTimestampedKey).Methods("GET")

	log.Println("key authority server started on port :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

// request a key from the key authority. We do not go over verification or authentication of key requests for demonstration purposes
func getKey(w http.ResponseWriter, r *http.Request) {
	attributes := r.URL.Query()["attribute"]
	fmt.Printf("generating key for attributes %v\n", attributes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scheme.KeyGen(attributes))
}

// request a key from the key authority that contains timestamp attributes
func getTimestampedKey(w http.ResponseWriter, r *http.Request) {
	attributes := r.URL.Query()["attribute"]
	fmt.Printf("generating timestamped key for attributes %v\n", attributes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scheme.KeyGen(append(attributes, generateTimestamp()...)))
}

// uptate the policy config entry in the database
func updatePolicyConfig() {

	writeKey := crypto.GenerateSignatureKey()

	newPolicyConfig := policyConfig.Config{
		PurposeTrees: utils.ExamplePurposeTrees(),
		Scheme:       crypto.ABEscheme{PublicKey: scheme.PublicKey},
	}

	createdTime := time.Now()
	uuid := utils.Assure(uuid.Parse(authorityUUID))

	publicKey := writeKey.PublicKey

	//curve is an interface type and can't be marshaled, we remove it and the database can add it back
	publicKey.Curve = nil
	marshaledPublicWriteKey := utils.ToBytes(publicKey)

	var checkSum bytes.Buffer

	// needed for MessagePack encoding
	/* 	for _, c := range newPolicyConfig.PurposeTrees {
		c.DisconnectParents()
	} */

	for _, s := range [][]byte{utils.ToBytes("relations"), uuid[:], []byte{}, marshaledPublicWriteKey, utils.ToBytes(newPolicyConfig), utils.ToBytes(createdTime)} {
		checkSum.Write(s)
	}

	signature := crypto.Sign(writeKey, checkSum.Bytes())

	newRecord := utils.Record{
		Table:           "relations",
		ID:              uuid,
		PrivateWriteKey: []byte{},
		PublicWriteKey:  marshaledPublicWriteKey,
		Data:            utils.ToBytes(newPolicyConfig),
		Created:         createdTime,
		Signature:       signature,
	}

	jsonData := utils.Assure(json.Marshal(newRecord))
	resp := utils.Assure(http.Post(databaseURL+"/entries", "application/json", bytes.NewBuffer(jsonData)))
	defer resp.Body.Close()

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("entry add failed: %s", body)
	}
}

// turn the current date to an array of attributes that represent the current timestamp
func generateTimestamp() []string {
	valueSize := 15
	value := time.Now().Unix()

	out := []string{}
	for i := valueSize - 1; i >= 0; i-- {
		// Shift and mask to get each bit
		bit := (value >> i) & 1
		out = append(out, strings.Repeat("*", valueSize-i-1)+fmt.Sprintf("%d", bit)+strings.Repeat("*", i))
	}
	return out
}
