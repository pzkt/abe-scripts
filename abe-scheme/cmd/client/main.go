package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/crypto"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
	pb "github.com/pzkt/abe-scripts/abe-scheme/proto"
	"github.com/pzkt/abe-scripts/generate-pseudodata/generator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	ctx          context.Context
	client       pb.RecordServiceClient
}

type Entry struct {
	ID       uuid.UUID
	Created  *timestamppb.Timestamp
	writeKey *ecdsa.PrivateKey
}

func main() {
	/* db := utils.Connect()
	defer db.Close() */

	env := setup()
	//defer env.conn.Close()
	//defer env.cancel()

	record := generator.GenerateCardiologyRecord("345")

	env.addEntry("table_one", record, "Phone AND (Analysis OR Purchase AND General-Purpose)", "Admin")

	fmt.Printf("%v", env.entries[0].ID)

	utils.Assure(env.getEntry(env.entries[0].ID.String()))

	//fmt.Println(generateBitAttributes(174897, 18))
	//out, _ := generateComparison(8, 4, Greater)
}

func setup() *env {
	conn := utils.Assure(grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials())))

	client := pb.NewRecordServiceClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)

	return &env{
		abeScheme:    crypto.Setup(),
		policyConfig: updatePolicyConfig(),
		entries:      []Entry{},
		client:       client,
		ctx:          ctx,
	}
}

func updatePolicyConfig() utils.PolicyConfig {
	//return example policy for now
	return utils.ExamplePolicyConfig()
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

	var newEntry Entry
	newEntry.writeKey = writeKey

	resp := utils.Assure(e.client.AddEntry(e.ctx, &pb.AddEntryRequest{
		Table:                   table,
		WriteKeyCipher:          writeKeyCipher,
		MarshaledPublicWriteKey: marshaledPublicWriteKey,
		DataCipher:              dataCipher,
	}))

	newEntry.Created = resp.GetCreated()
	newEntry.ID = utils.Assure(uuid.Parse(resp.GetId()))

	e.entries = append(e.entries, newEntry)
}

func (e *env) getEntry(recordID string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.GetEntry(ctx, &pb.GetEntryRequest{Id: recordID})
	if err != nil {
		return nil, fmt.Errorf("RPC failed: %v", err)
	}

	return resp.Data, nil
}

func modifyEntry(table string) {

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
