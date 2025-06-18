package main

import (
	"fmt"
	"strings"

	"github.com/fentec-project/gofe/abe"
	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

type Ops int

const (
	Less Ops = iota
	Greater
	LessOrEqual
	GreaterOrEqual
)

type LargeData struct {
	WriteKey string `db:"jsonb"`
	Data     string `db:"jsonb"`
}

func main() {
	//db := utils.Connect()
	//defer db.Close()

	curPolicy := updatePolicyConfig()
	toAttr("(Direct AND Analysis) OR Masked-Research", curPolicy)
	//fmt.Println(generateBitAttributes(174897, 18))

}

func updatePolicyConfig() utils.PolicyConfig {
	//return example policy for now
	return utils.ExamplePolicyConfig()
}

func addEntry(table string, id string, entry any, policy string, purposes string, policyConfig utils.PolicyConfig) {

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

func encryptFile(path string, pubKey *abe.FAMEPubKey) {

}

func generateBitAttributes(value uint, valueSize int) []string {
	out := []string{}
	for i := valueSize - 1; i >= 0; i-- {
		// Shift and mask to get each bit
		bit := (value >> i) & 1
		out = append(out, strings.Repeat("*", valueSize-i-1)+fmt.Sprintf("%d", bit)+strings.Repeat("*", i))
	}
	return out
}

func generateComparison(value int, valueSize int, op Ops) (string, error) {
	switch op {
	case GreaterOrEqual:
		return generateComparison(value-1, valueSize, Greater)
	case LessOrEqual:
		return generateComparison(value+1, valueSize, Less)
	}

	gates := [2]string{" AND ", " OR "}
	out := ""

	for i := valueSize - 1; i > 0; i-- {
		bit := (value >> i) & 1
		switch bit {
		case 0:
			mask := (1 << (i)) - 1
			if op == Greater && ^(mask&value)&mask == 0 {
				out += strings.Repeat("*", valueSize-i-1) + fmt.Sprintf("%d", op) + strings.Repeat("*", i)
				return out, nil
			}
			out += strings.Repeat("*", valueSize-i-1) + fmt.Sprintf("%d", op) + strings.Repeat("*", i) + gates[op]
		case 1:
			mask := (1 << (i)) - 1
			if op == Less && mask&value == 0 {
				out += strings.Repeat("*", valueSize-i-1) + fmt.Sprintf("%d", op) + strings.Repeat("*", i)
				return out, nil
			}
			out += strings.Repeat("*", valueSize-i-1) + fmt.Sprintf("%d", op) + strings.Repeat("*", i) + gates[1-op]
		}
	}
	out += strings.Repeat("*", valueSize-1) + fmt.Sprintf("%d", op)
	return out, nil
}
