package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
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

type Record struct {
	Table           string    `json:"table"`
	ID              string    `json:"id"`
	PrivateWriteKey []byte    `json:"private_write_key"`
	PublicWriteKey  []byte    `json:"public_write_key"`
	Data            []byte    `json:"data"`
	Created         time.Time `json:"created"`
}

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

func addEntry(w http.ResponseWriter, r *http.Request) {
	var record Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

	fmt.Printf("new entry added in table: %s with uuid: %s\n", record.Table, record.ID)
}

func getEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	table := vars["table"]
	id := vars["id"]

	var record Record
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

func getWriteKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	table := vars["table"]
	id := vars["id"]

	var record Record
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
