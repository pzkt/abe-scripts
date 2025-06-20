package main

import (
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/crypto"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"

	"github.com/pzkt/abe-scripts/generate-pseudodata/generator"
)

type Ops int

const (
	Less Ops = iota
	Greater
	LessOrEqual
	GreaterOrEqual
)

type env struct {
	ABEscheme    *crypto.ABEscheme
	DB           *sql.DB
	policyConfig utils.PolicyConfig
}

func main() {

	/* 	key := crypto.GenerateSignatureKey()

	   	test := "this is a message"

	   	signed := crypto.Sign(key, utils.ToBytes(test))

	   	fmt.Println("signed: ", signed)
	   	fmt.Println("correct: ", crypto.Verify(&key.PublicKey, utils.ToBytes(test), signed))
	   	return */

	db := utils.Connect()
	defer db.Close()

	env := setup(db)

	record := generator.GenerateCardiologyRecord("345")
	fmt.Printf("%+v\n", record)

	env.addEntry("table_one", record, "Phone AND (Analysis OR Purchase AND General-Purpose)", "Admin")
	//fmt.Println(generateBitAttributes(174897, 18))
	//out, _ := generateComparison(8, 4, Greater)
}

func setup(db *sql.DB) *env {
	return &env{
		ABEscheme:    crypto.Setup(),
		DB:           db,
		policyConfig: updatePolicyConfig(),
	}
}

func updatePolicyConfig() utils.PolicyConfig {
	//return example policy for now
	return utils.ExamplePolicyConfig()
}

func (e *env) addEntry(table string, entry any, readPurposes string, writePurposes string) {

	fullReadPurposes := toAttr(readPurposes, e.policyConfig)
	fullWritePurposes := toAttr(writePurposes, e.policyConfig)

	/* 	combinedPolicy := []string{}
	   	for _, str := range []string{fullReadPurposes} {
	   		if str != "" {
	   			combinedPolicy = append(combinedPolicy, str)
	   		}
	   	}

	   	fullPolicy := strings.Join(combinedPolicy, " AND ") */

	fmt.Println("read policy: " + fullReadPurposes)
	fmt.Println("write policy: " + fullWritePurposes)

	writeKey := crypto.GenerateSignatureKey()

	dataCipher := e.ABEscheme.Encrypt(utils.ToBytes(entry), fullReadPurposes)

	//custom marshal functions for elliptic curve keys
	marshaledWriteKey := utils.Assure(x509.MarshalECPrivateKey(writeKey))
	marshaledPublicWriteKey := utils.Assure(writeKey.PublicKey.ECDH()).Bytes()

	writeKeyCipher := e.ABEscheme.Encrypt(marshaledWriteKey, fullWritePurposes)

	// !!! SQL INJECTION RISK (But that's fine for demonstration purposes) !!!
	query := fmt.Sprintf(`INSERT INTO %s (private_write_key, public_write_key, data) VALUES ($1, $2, $3)`, table)
	_, err := e.DB.Exec(
		query,
		writeKeyCipher,
		marshaledPublicWriteKey,
		dataCipher,
	)
	if err != nil {
		log.Fatalf("Insert failed: %v", err)
	}
}

func modifyEntry(table string) {

}

func getEntry() {

}

func getRow() {

}

func getTransformRow() {

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
