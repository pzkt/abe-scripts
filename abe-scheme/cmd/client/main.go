package main

import (
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type Ops int

const (
	Less Ops = iota
	Greater
	LessOrEqual
	GreaterOrEqual
)

func main() {
	res, _ := generateComparison(10, 4, Greater)
	println(res)
	//db := utils.Connect()
	//fmt.Println("client")
	//db.Close()
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

func generateComparison(value uint, valueSize int, op Ops) (string, error) {
	switch op {
	case GreaterOrEqual:
		generateComparison(value-1, valueSize, Greater)
	case LessOrEqual:
		generateComparison(value+1, valueSize, Less)
	case Greater:
	case Less:
	}

	//LESS
	gates := [2]string{" AND ", " OR "}
	out := ""

	for i := valueSize - 1; i > 0; i-- {
		// Shift and mask to get each bit
		bit := (value >> i) & 1
		switch bit {
		case 0:
			out += strings.Repeat("*", valueSize-i-1) + fmt.Sprintf("%d", op) + strings.Repeat("*", i) + gates[op]
		case 1:
			out += strings.Repeat("*", valueSize-i-1) + fmt.Sprintf("%d", op) + strings.Repeat("*", i) + gates[1-op]
		}
	}
	out += strings.Repeat("*", valueSize-1) + fmt.Sprintf("%d", op)
	return out, nil
}
