package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// generic struct for transmitting information between different parties
type Record struct {
	Table           string    `json:"table"`
	ID              string    `json:"id"`
	PrivateWriteKey []byte    `json:"private_write_key"`
	PublicWriteKey  []byte    `json:"public_write_key"`
	Data            []byte    `json:"data"`
	Created         time.Time `json:"created"`
	Signature       []byte    `json:"signature"`
}

// try will exit if the function returned an error
func Try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// assure will return the value of a function with return type (any, error) if no error ocurred
// else it will exit the program
func Assure[A any](result A, err error) A {
	if err != nil {
		log.Fatal(err)
	}
	return result
}

// connects to the postgres database and returns an sql.DB variable
func Connect() *sql.DB {
	db_password := "pwd"
	connection := fmt.Sprintf("postgres://postgres:%s@localhost:5432/data?sslmode=disable", db_password)

	db := Assure(sql.Open("postgres", connection))
	Try(db.Ping())
	return db
}

// turn anything into bytes
func ToBytes(a any) []byte {
	return Assure(json.Marshal(a))
}

// turn the bytes from ToBytes back to a struct (pass the struct as a pointer)
func FromBytes(data []byte, target any) {
	Try(json.Unmarshal(data, target))
}
