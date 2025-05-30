package utils

import (
	"database/sql"
	"fmt"
	"log"
)

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
