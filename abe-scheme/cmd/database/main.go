package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

//docker run --name postgres-container -e POSTGRES_PASSWORD=pwd -p 5432:5432 -d postgres

type Product struct {
	Name      string
	Price     float64
	Available bool
}

func main() {
	dbPassword := "pwd"
	connection := fmt.Sprintf("postgres://postgres:%s@localhost:5432/data?sslmode=disable", dbPassword)

	db, err := sql.Open("postgres", connection)

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	createTable(db)

	product := Product{"Boook", 15.55, true}
	pk := insertProduct(db, product)

	var name string
	query := "SELECT name FROM product WHERE id=$1"
	err = db.QueryRow(query, pk).Scan(&name)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s\n", name)
}

func setup(db *sql.DB) {
	//create the key-value table for table row relations
	query := `CREATE TABLE IF NOT EXISTS relations (
		id SERIAL PRIMARY KEY,
		write_key JSONB,
		data JSONB,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))
}

func createTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS product (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		price NUMERIC(6,2) NOT NULL,
		available BOOLEAN,
		created timestamp DEFAULT NOW()
	)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertProduct(db *sql.DB, product Product) int {
	query := `INSERT INTO product (name, price, available)
		VALUES ($1, $2, $3) RETURNING id` //$1 can prevent SQL injection

	var pk int
	err := db.QueryRow(query, product.Name, product.Price, product.Available).Scan(&pk)
	if err != nil {
		log.Fatal(err)
	}
	return pk
}
