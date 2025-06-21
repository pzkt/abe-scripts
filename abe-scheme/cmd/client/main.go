package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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
	abeScheme    *crypto.ABEscheme
	policyConfig utils.PolicyConfig
	entries      []Entry
}

type Entry struct {
	ID       uuid.UUID
	Created  time.Time
	writeKey *ecdsa.PrivateKey
}

type Record struct {
	Table           string    `json:"table"`
	ID              string    `json:"id"`
	PrivateWriteKey []byte    `json:"private_write_key"`
	PublicWriteKey  []byte    `json:"public_write_key"`
	Data            []byte    `json:"data"`
	Created         time.Time `json:"created"`
}

const databaseURL = "http://localhost:8080"
const authorityURL = "http://localhost:8081"

func main() {
	/* db := utils.Connect()
	defer db.Close() */

	env := setup()
	//defer env.conn.Close()
	//defer env.cancel()

	cipher := env.abeScheme.Encrypt(utils.ToBytes("wow schgloopy"), "test AND wow")

	newKey := requestNewKey([]string{"test", "wow"})

	fmt.Printf("%+v\n", env.abeScheme.Decrypt())

	return

	record := generator.GenerateCardiologyRecord("345")
	env.addEntry("table_one", record, "Phone AND (Analysis OR Purchase AND General-Purpose)", "Admin")
	fmt.Printf("%v", env.entries[0].ID)

	fmt.Printf("%+v", env.getEntry("table_one", env.entries[0].ID.String()))

	//fmt.Println(generateBitAttributes(174897, 18))
	//out, _ := generateComparison(8, 4, Greater)
}

func setup() *env {
	return &env{
		abeScheme:    crypto.Setup(),
		policyConfig: updatePolicyConfig(),
		entries:      []Entry{},
	}
}

func updatePolicyConfig() utils.PolicyConfig {
	//return example policy for now
	return utils.ExamplePolicyConfig()
}

func requestNewKey(attributes []string) []byte {
	req := utils.Assure(http.NewRequest("GET", authorityURL+"/get_key", nil))

	q := req.URL.Query()
	for _, attr := range attributes {
		q.Add("attribute", attr)
	}
	req.URL.RawQuery = q.Encode()

	resp := utils.Assure(http.DefaultClient.Do(req))
	defer resp.Body.Close()

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("request new key failed: %s", body)
	}

	key := []byte{}
	utils.Try(json.Unmarshal(body, &key))
	return key
}

func (e *env) addEntry(table string, entry any, readPurposes string, writePurposes string) {

	fullReadPurposes := toAttr(readPurposes, e.policyConfig)
	fullWritePurposes := toAttr(writePurposes, e.policyConfig)

	writeKey := crypto.GenerateSignatureKey()

	dataCipher := e.abeScheme.Encrypt(utils.ToBytes(entry), fullReadPurposes)

	//custom marshal functions for elliptic curve keys
	marshaledWriteKey := utils.Assure(x509.MarshalECPrivateKey(writeKey))
	marshaledPublicWriteKey := utils.Assure(writeKey.PublicKey.ECDH()).Bytes()

	writeKeyCipher := e.abeScheme.Encrypt(marshaledWriteKey, fullWritePurposes)

	createdTime := time.Now()
	newUUID := uuid.New()

	newRecord := Record{
		Table:           table,
		ID:              newUUID.String(),
		PrivateWriteKey: writeKeyCipher,
		PublicWriteKey:  marshaledPublicWriteKey,
		Data:            dataCipher,
		Created:         createdTime,
	}

	jsonData := utils.Assure(json.Marshal(newRecord))
	resp := utils.Assure(http.Post(databaseURL+"/entries", "application/json", bytes.NewBuffer(jsonData)))
	defer resp.Body.Close()

	var newEntry = Entry{
		writeKey: writeKey,
		Created:  createdTime,
		ID:       newUUID,
	}

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("entry add failed: %s", body)
	}

	e.entries = append(e.entries, newEntry)
}

func (e *env) getEntry(table string, recordID string) Record {
	resp := utils.Assure(http.Get(fmt.Sprintf("%s/entries/%s/%s", databaseURL, table, recordID)))
	defer resp.Body.Close()

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("get entry failed: %s", body)
	}

	var record Record
	utils.Try(json.Unmarshal(body, &record))
	return record
}

func modifyEntry(table string) {

}

func getRow() {

}

func getTransformRow() {

}

func (e *env) getWriteKey(table string, recordID string) Record {
	resp := utils.Assure(http.Get(fmt.Sprintf("%s/write_key/%s/%s", databaseURL, table, recordID)))
	defer resp.Body.Close()

	fmt.Println(fmt.Sprintf("%s/write_key/%s/%s", databaseURL, table, recordID))

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("get write key failed: %s", body)
	}

	var record Record
	utils.Try(json.Unmarshal(body, &record))
	return record
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
