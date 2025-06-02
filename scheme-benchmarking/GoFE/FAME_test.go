package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/csv"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/fentec-project/gofe/abe"
)

var attributeCounts = [...]int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50}
var msg = "Svx7QqFWUqDJ6hOo4dByAGqmXOUNOeGP"

func updateCSV(fileName, index, column, value string) error {
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

func EncryptAES(key []byte, plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return ciphertext, nil
}

func BenchmarkSetup(b *testing.B) {
	for n := 0; n < b.N; n++ {
		a := abe.NewFAME()
		pubKey, secKey, _ := a.GenerateMasterKeys()
		runtime.KeepAlive(pubKey)
		runtime.KeepAlive(secKey)
	}
}

func BenchmarkKeyGen(b *testing.B) {
	a := abe.NewFAME()
	_, secKey, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			gamma := make([]string, count)
			for a := 0; a < count; a++ {
				gamma[a] = fmt.Sprintf("attribute_%d", a)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				keys, _ := a.GenerateAttribKeys(gamma, secKey)
				runtime.KeepAlive(keys)
			}
		})
	}
}

func BenchmarkEncryptionAND(b *testing.B) {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " AND ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				msp, _ := abe.BooleanToMSP(policy.String(), false)
				cipher, _ := a.Encrypt(msg, msp, pubKey)
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func BenchmarkEncryptionOR(b *testing.B) {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " OR ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				msp, _ := abe.BooleanToMSP(policy.String(), false)
				cipher, _ := a.Encrypt(msg, msp, pubKey)
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func getSize(s any) int {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	err := enc.Encode(s)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	return len(network.Bytes())
}

func TestCipherSizeAND(t *testing.T) {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()

	for size := range 25 {
		content := make([]byte, 1<<size)
		rand.Read(content)

		for _, count := range attributeCounts {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " AND ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)

			cipher, _ := a.Encrypt(msg, msp, pubKey)
			aes_cipher, _ := EncryptAES([]byte(msg), string(content))
			pure_abe, _ := a.Encrypt(string(content), msp, pubKey)

			_ = cipher

			updateCSV("gofe_fame_ciphertext_AND.csv", fmt.Sprint(count), "single "+fmt.Sprint(1<<size), fmt.Sprint(getSize(pure_abe)))
			updateCSV("gofe_fame_ciphertext_AND.csv", fmt.Sprint(count), "hybrid "+fmt.Sprint(1<<size), fmt.Sprint(len(aes_cipher)+getSize(cipher)))
		}
	}
}

func TestCipherSizeOR(t *testing.T) {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()

	for size := range 25 {
		content := make([]byte, 1<<size)
		rand.Read(content)

		for _, count := range attributeCounts {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " OR ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)

			cipher, _ := a.Encrypt(msg, msp, pubKey)
			aes_cipher, _ := EncryptAES([]byte(msg), string(content))
			pure_abe, _ := a.Encrypt(string(content), msp, pubKey)

			_ = cipher

			updateCSV("gofe_fame_ct.csv", fmt.Sprint(count), "single "+fmt.Sprint(1<<size), fmt.Sprint(getSize(pure_abe)))
			updateCSV("gofe_fame_ct.csv", fmt.Sprint(count), "hybrid "+fmt.Sprint(1<<size), fmt.Sprint(len(aes_cipher)+getSize(cipher)))
		}
	}
}

func BenchmarkDecryptionOR(b *testing.B) {
	a := abe.NewFAME()
	pubKey, secKey, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " OR ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)
			cipher, _ := a.Encrypt(msg, msp, pubKey)

			keys, _ := a.GenerateAttribKeys([]string{fmt.Sprintf("attribute_%d", count-1)}, secKey)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				text, e := a.Decrypt(cipher, keys, pubKey)
				if e != nil {
					panic(e)
				}
				runtime.KeepAlive(text)
			}
		})
	}
}

func BenchmarkDecryptionAND(b *testing.B) {
	a := abe.NewFAME()
	pubKey, secKey, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " AND ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)
			cipher, _ := a.Encrypt(msg, msp, pubKey)

			gamma := make([]string, count)
			for a := 0; a < count; a++ {
				gamma[a] = fmt.Sprintf("attribute_%d", a)
			}
			keys, _ := a.GenerateAttribKeys(gamma, secKey)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				text, e := a.Decrypt(cipher, keys, pubKey)
				if e != nil {
					panic(e)
				}
				runtime.KeepAlive(text)
			}
		})
	}
}

func main() {
	res := testing.Benchmark(BenchmarkEncryptionOR)
	fmt.Printf("%s\n%#[1]v\n", res)
}
