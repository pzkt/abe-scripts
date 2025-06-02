package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

//docker run --name postgres-container -e POSTGRES_PASSWORD=pwd -p 5432:5432 -d postgres

//docker start -a postgres-container

//docker run --name pgadmin -p 15432:80 -e 'PGADMIN_DEFAULT_EMAIL=user@domain.com' -e 'PGADMIN_DEFAULT_PASSWORD=pwd' -d dpage/pgadmin4

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

	setup(db)
	createTable(db, "cardiology_records", utils.CardiologyRecord{})

	entry := utils.CardiologyRecord{Notes: "test", BloodPressure: 80, HeartRate: 20, StressTestResults: "good", CardiacMedications: "none", EFPercentage: 80}

	utils.Try(AddEntry(db, "cardiology_records", entry))

	/*
		 	product := Product{"Boook", 15.55, true}
			createTable(db, "product", product)

			pk := insertProduct(db, product)

			var name string
			query := "SELECT name FROM product WHERE id=$1"
			err = db.QueryRow(query, pk).Scan(&name)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Name: %s\n", name)
	*/
}

func setup(db *sql.DB) {
	//create the key-value table for table row relations
	query := `CREATE TABLE IF NOT EXISTS relations (
		id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		write_key JSONB,
		data JSONB,
		created TIMESTAMP DEFAULT NOW()
	)`

	utils.Assure(db.Exec(query))
}

/* func insertProduct(db *sql.DB, product Product) int {
	query := `INSERT INTO product (name, price, available)
		VALUES ($1, $2, $3) RETURNING id` //$1 can prevent SQL injection

	var pk int
	err := db.QueryRow(query, product.Name, product.Price, product.Available).Scan(&pk)
	if err != nil {
		log.Fatal(err)
	}
	return pk
} */

func createTable(db *sql.DB, tableName string, model any) error {
	// Get the type of the model
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct or pointer to struct")
	}

	var columns []string
	columns = append(columns, "id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY")

	// Iterate through the struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columnName := strings.ToLower(field.Name)

		if slices.Contains([]string{"id"}, columnName) {
			return fmt.Errorf("reserved column name used")
		}

		columnType := ""

		// Map Go types to PostgreSQL types
		switch field.Type.Kind() {
		case reflect.String:
			columnType = "TEXT"
		case reflect.Int, reflect.Int32:
			columnType = "INTEGER"
		case reflect.Int64:
			columnType = "BIGINT"
		case reflect.Float32:
			columnType = "REAL"
		case reflect.Float64:
			columnType = "DOUBLE PRECISION"
		case reflect.Bool:
			columnType = "BOOLEAN"
		case reflect.Struct:
			if field.Type.String() == "time.Time" {
				columnType = "TIMESTAMP"
			}
		default:
			// Skip unsupported types
			continue
		}

		columns = append(columns, fmt.Sprintf("%s %s", columnName, columnType))
	}

	if len(columns) == 0 {
		return fmt.Errorf("no valid columns found in struct")
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n);", tableName, strings.Join(columns, ",\n\t"))

	_, err := db.Exec(query)
	return err
}

func AddEntry(db *sql.DB, tableName string, data any) error {
	dataType := reflect.TypeOf(data)
	dataValue := reflect.ValueOf(data)

	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
		dataValue = dataValue.Elem()
	}

	if dataType.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct or pointer to struct")
	}

	columns, err := getTableColumns(db, tableName)
	if err != nil {
		return fmt.Errorf("failed to get table columns: %v", err)
	}

	// Prepare data for insertion
	var fieldNames []string
	var placeholders []string
	var values []interface{}
	fieldCount := 0

	// Iterate through struct fields
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		fieldValue := dataValue.Field(i)

		// Get column name from struct tag or field name
		columnName := strings.ToLower(field.Name)
		if tag := field.Tag.Get("db"); tag != "" {
			columnName = tag
		}

		// Skip if column doesn't exist in table
		if !slices.Contains(columns, columnName) {
			continue
		}

		fieldNames = append(fieldNames, columnName)
		placeholders = append(placeholders, fmt.Sprintf("$%d", fieldCount+1))
		values = append(values, fieldValue.Interface())
		fieldCount++
	}

	if len(fieldNames) == 0 {
		return errors.New("no matching fields found between struct and table")
	}

	if len(fieldNames) != len(columns)-1 {
		return errors.New("not all fields in struct could be linked to columns")
	}

	// Build the SQL query
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(fieldNames, ", "),
		strings.Join(placeholders, ", "),
	)

	fmt.Println("inserting new data")
	_, err = db.Exec(query, values...)
	return err
}

func getTableColumns(db *sql.DB, tableName string) ([]string, error) {
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = $1
	`

	rows := utils.Assure(db.Query(query, tableName))

	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("table '%s' not found or has no columns", tableName)
	}

	return columns, nil
}
