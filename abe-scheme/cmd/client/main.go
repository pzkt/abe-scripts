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
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils/policyConfig"
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
	policyConfig policyConfig.Config
	entries      map[uuid.UUID]Entry
}

type Entry struct {
	Created  time.Time
	writeKey *ecdsa.PrivateKey
}

const databaseURL = "http://localhost:8080"
const authorityURL = "http://localhost:8081"

const authorityUUID = "497dcba3-ecbf-4587-a2dd-5eb0665e6880"

func main() {
	/* 	key := crypto.GenerateSignatureKey()
	   	publicKey := key.PublicKey

	   	publicKey.Curve = nil
	   	byteKey := utils.ToBytes(publicKey)

	   	var newKey ecdsa.PublicKey
	   	utils.FromBytes(byteKey, &newKey)

	   	data := []byte{45, 213, 43, 6, 43, 3}

	   	signature := crypto.Sign(key, data)

	   	newKey.Curve = elliptic.P256()
	   	fmt.Println(crypto.Verify(&newKey, data, signature))

	   	return */

	env := setup()

	ABEkey := requestNewKey([]string{"Admin"})

	record := generator.GenerateCardiologyRecord("345")

	addedUUID := env.addEntry("table_one", record, "Profiling OR Marketing", "Admin")

	fmt.Println("first plaintext")
	ciphertext := env.getEntry("table_one", addedUUID).Data
	fmt.Println(string(env.abeScheme.Decrypt(ciphertext, ABEkey)))

	record.PatientID = "wow schgloopy"
	env.modifyEntry("table_one", record, "Profiling OR Marketing", "Admin", addedUUID)

	fmt.Println("second plaintext")
	ciphertext = env.getEntry("table_one", addedUUID).Data
	fmt.Println(string(env.abeScheme.Decrypt(ciphertext, ABEkey)))

	//fmt.Println(generateBitAttributes(174897, 18))
	//out, _ := generateComparison(8, 4, Greater)
}

func setup() *env {
	newEnv := env{
		abeScheme: crypto.Setup(),
		entries:   make(map[uuid.UUID]Entry),
	}

	newEnv.updatePolicyConfig()
	return &newEnv
}

func (e *env) updatePolicyConfig() {
	data := e.getEntry("relations", utils.Assure(uuid.Parse(authorityUUID))).Data
	utils.FromBytes(data, &e.policyConfig)

	//we need to reconnect the parents because serialization forces us to remove cyclical references
	for _, tree := range e.policyConfig.PurposeTrees {
		tree.ReconnectParents(nil)
	}

	e.abeScheme.PublicKey = e.policyConfig.Scheme.PublicKey
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

func (e *env) addEntry(table string, entry any, readPurposes string, writePurposes string) uuid.UUID {
	newUUID := uuid.New()
	e.modifyEntry(table, entry, readPurposes, writePurposes, newUUID)
	return newUUID
}

func (e *env) modifyEntry(table string, entry any, readPurposes string, writePurposes string, newUUID uuid.UUID) {

	fullReadPurposes := toAttr(readPurposes, e.policyConfig)
	fullWritePurposes := toAttr(writePurposes, e.policyConfig)

	writeKey := crypto.GenerateSignatureKey()

	dataCipher := e.abeScheme.Encrypt(utils.ToBytes(entry), fullReadPurposes)

	//custom marshal functions for elliptic curve keys
	marshaledWriteKey := utils.Assure(x509.MarshalECPrivateKey(writeKey))
	publicKey := writeKey.PublicKey

	//curve is an interface type and can't be marshaled, we remove it and the database can add it back
	publicKey.Curve = nil
	marshaledPublicWriteKey := utils.ToBytes(publicKey)

	writeKeyCipher := e.abeScheme.Encrypt(marshaledWriteKey, fullWritePurposes)

	createdTime := time.Now()

	//prevent any part of the record to be tampered with by using all parts to generate the signature
	var checkSum bytes.Buffer
	for _, s := range [][]byte{utils.ToBytes(table), newUUID[:], writeKeyCipher, marshaledPublicWriteKey, dataCipher, utils.ToBytes(createdTime)} {
		checkSum.Write(s)
	}

	signature := crypto.Sign(writeKey, checkSum.Bytes())

	newRecord := utils.Record{
		Table:           table,
		ID:              newUUID,
		PrivateWriteKey: writeKeyCipher,
		PublicWriteKey:  marshaledPublicWriteKey,
		Data:            dataCipher,
		Created:         createdTime,
		Signature:       signature,
	}

	jsonData := utils.Assure(json.Marshal(newRecord))
	resp := utils.Assure(http.Post(databaseURL+"/entries", "application/json", bytes.NewBuffer(jsonData)))
	defer resp.Body.Close()

	var newEntry = Entry{
		writeKey: writeKey,
		Created:  createdTime,
	}

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("entry add failed: %s", body)
	}

	e.entries[newUUID] = newEntry
}

func (e *env) getEntry(table string, recordID uuid.UUID) utils.Record {
	resp := utils.Assure(http.Get(fmt.Sprintf("%s/entries/%s/%s", databaseURL, table, recordID)))
	defer resp.Body.Close()

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("get entry failed: %s", body)
	}

	var record utils.Record
	utils.Try(json.Unmarshal(body, &record))
	return record
}

func getRow() {

}

func (e *env) getWriteKey(table string, recordID string) utils.Record {
	resp := utils.Assure(http.Get(fmt.Sprintf("%s/write_key/%s/%s", databaseURL, table, recordID)))
	defer resp.Body.Close()

	fmt.Println(fmt.Sprintf("%s/write_key/%s/%s", databaseURL, table, recordID))

	body := utils.Assure(io.ReadAll(resp.Body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("get write key failed: %s", body)
	}

	var record utils.Record
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
