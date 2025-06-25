package utils

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

// generic struct for transmitting information between different parties
type Record struct {
	Table           string    `json:"table"`
	ID              uuid.UUID `json:"id"`
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
	return ToBytesCbor(a)
}

// turn the bytes from ToBytes back to a struct (pass the struct as a pointer)
func FromBytes(data []byte, target any) {
	FromBytesCbor(data, target)
}

// --- byte encoding using msgPack ---

func ToBytesMsgPack(a any) []byte {
	return Assure(msgpack.Marshal(a))
}

func FromBytesMsgPack(data []byte, target any) {
	Try(msgpack.Unmarshal(data, target))
}

// --- byte encoding using naive json ---

func ToBytesJson(a any) []byte {
	return Assure(json.Marshal(a))
}

func FromBytesJson(data []byte, target any) {
	Try(json.Unmarshal(data, target))
}

// --- byte encoding using cbor ---

func ToBytesCbor(a any) []byte {
	return Assure(cbor.Marshal(a))
}

func FromBytesCbor(data []byte, target any) {
	Try(cbor.Unmarshal(data, target))
}

// helper function for writing into a CSV file
func UpdateCSV(fileName, index, column, value string) error {
	var records [][]string
	var headers []string
	fileExists := true

	file, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fileExists = false
		} else {
			return fmt.Errorf("error opening file: %w", err)
		}
	}

	if fileExists {
		defer file.Close()
		reader := csv.NewReader(file)
		records, err = reader.ReadAll()
		if err != nil {
			return fmt.Errorf("error reading CSV: %w", err)
		}

		if len(records) > 0 {
			headers = records[0]
		}
	}

	if len(records) == 0 {
		headers = []string{"index"}
		records = [][]string{headers}
	}

	columnIndex := -1
	for i, h := range headers {
		if h == column {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		headers = append(headers, column)
		columnIndex = len(headers) - 1
		records[0] = headers // Update header row

		for i := 1; i < len(records); i++ {
			if len(records[i]) <= columnIndex {
				for len(records[i]) < columnIndex {
					records[i] = append(records[i], "")
				}
				records[i] = append(records[i], "")
			}
		}
	}

	rowIndex := -1
	for i, record := range records {
		if i == 0 {
			continue // Skip header row
		}
		if len(record) > 0 && record[0] == index {
			rowIndex = i
			break
		}
	}

	if rowIndex == -1 {
		newRow := make([]string, len(headers))
		newRow[0] = index
		for i := 1; i < len(newRow); i++ {
			newRow[i] = ""
		}
		records = append(records, newRow)
		rowIndex = len(records) - 1
	}

	for len(records[rowIndex]) <= columnIndex {
		records[rowIndex] = append(records[rowIndex], "")
	}

	records[rowIndex][columnIndex] = value
	file, err = os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.WriteAll(records)
	if err != nil {
		return fmt.Errorf("error writing CSV: %w", err)
	}

	return nil
}
