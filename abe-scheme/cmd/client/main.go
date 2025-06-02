package main

import (
	"fmt"

	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"

	_ "github.com/lib/pq"
)

func main() {
	db := utils.Connect()
	fmt.Println("client")
	db.Close()
}

func addEntry(table string, id string, entry any) {

}

func modifyEntry(table string, id string) {

}

func modifyPolicy(fieldId string) {

}

func getEntry() {

}

func getRow() {

}

func getTransformRow() {

}
