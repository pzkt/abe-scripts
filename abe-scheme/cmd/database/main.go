/*

The database is the only application that should have access to the postgreSQL database.
They must validate incoming modification requests

*/

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/crypto"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

/*
docker setup commands:

docker run --name postgres-container -e POSTGRES_PASSWORD=pwd -p 5432:5432 -d postgres

docker start -a postgres-container

docker run --name pgadmin -p 15432:80 -e 'PGADMIN_DEFAULT_EMAIL=user@domain.com' -e 'PGADMIN_DEFAULT_PASSWORD=pwd' -d dpage/pgadmin4

Host name/address: 172.17.0.2
Port: 5432
Maintenance database: postgres
Username: postgres
Password: pwd

*/

var db *sql.DB

func main() {
	dbPassword := "pwd"
	connection := fmt.Sprintf("postgres://postgres:%s@localhost:5432/data?sslmode=disable", dbPassword)

	db = utils.Assure(sql.Open("postgres", connection))
	utils.Try(db.Ping())

	defer db.Close()

	setup(db)

	r := mux.NewRouter()
	r.HandleFunc("/entries", addEntry).Methods("POST")
	r.HandleFunc("/entries/{table}/{id}", getEntry).Methods("GET")
	r.HandleFunc("/write_key/{table}/{id}", getWriteKey).Methods("GET")

	log.Println("database server started on port :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func setup(db *sql.DB) {
	//create the key-value table for table row relations
	query := `CREATE TABLE IF NOT EXISTS relations (
		id UUID,
		private_write_key BYTEA,
		public_write_key BYTEA,
		data BYTEA,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))

	query = `CREATE TABLE IF NOT EXISTS table_one (
		id UUID,
		private_write_key BYTEA,
		public_write_key BYTEA,
		data BYTEA,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))

	query = `CREATE TABLE IF NOT EXISTS table_two (
		id UUID,
		private_write_key BYTEA,
		public_write_key BYTEA,
		data BYTEA,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))
}

// add entry or validate if the given UUID already exists
func addEntry(w http.ResponseWriter, r *http.Request) {
	var record utils.Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var exists bool
	existQuery := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)`, record.Table)
	utils.Try(db.QueryRow(existQuery, record.ID).Scan(&exists))

	if exists {
		var oldRecord utils.Record
		getQuery := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, record.Table)
		utils.Try(db.QueryRow(getQuery, record.ID).Scan(&oldRecord.ID, &oldRecord.PrivateWriteKey, &oldRecord.PublicWriteKey, &oldRecord.Data, &oldRecord.Created))

		var checkSum bytes.Buffer
		for _, s := range [][]byte{utils.ToBytes(record.Table), record.ID[:], record.PrivateWriteKey, record.PublicWriteKey, record.Data, utils.ToBytes(record.Created)} {
			checkSum.Write(s)
		}

		var publicKey ecdsa.PublicKey
		utils.FromBytes(record.PublicWriteKey, &publicKey)
		publicKey.Curve = elliptic.P256()

		valid := crypto.Verify(&publicKey, checkSum.Bytes(), record.Signature)

		if !valid {
			fmt.Printf("Signature mismatch: modify request rejected!\n")
			return
		}
		fmt.Printf("Signature verified: modifying entry in table: %s with uuid: %s\n", record.Table, record.ID)

	} else {
		fmt.Printf("creating new entry in table: %s with uuid: %s\n", record.Table, record.ID)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (id, private_write_key, public_write_key, data, created) 
         VALUES ($1, $2, $3, $4, $5) 
		 ON CONFLICT (id) DO UPDATE SET
		 private_write_key = EXCLUDED.private_write_key,
		 public_write_key = EXCLUDED.public_write_key,
		 data = EXCLUDED.data,
		 created = EXCLUDED.created`,
		record.Table,
	)

	utils.Assure(db.Exec(query,
		record.ID,
		record.PrivateWriteKey,
		record.PublicWriteKey,
		record.Data,
		record.Created,
	))
}

// return the data field of an entry
func getEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	table := vars["table"]
	id := vars["id"]

	var record utils.Record
	query := fmt.Sprintf(`SELECT data, created FROM %s WHERE id = $1`, table)
	err := db.QueryRow(query, id).Scan(&record.Data, &record.Created)

	if err == sql.ErrNoRows {
		http.Error(w, "record not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// return the private write key field of an entry
func getWriteKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	table := vars["table"]
	id := vars["id"]

	var record utils.Record
	query := fmt.Sprintf(`SELECT private_write_key FROM %s WHERE id = $1`, table)
	err := db.QueryRow(query, id).Scan(&record.PrivateWriteKey)

	if err == sql.ErrNoRows {
		http.Error(w, "record not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}
