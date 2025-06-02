package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/cloudflare/circl/abe/cpabe/tkn20"
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
		pubKey, secKey, _ := tkn20.Setup(rand.Reader)
		runtime.KeepAlive(pubKey)
		runtime.KeepAlive(secKey)
	}
}

func BenchmarkKeyGen(b *testing.B) {
	_, secKey, _ := tkn20.Setup(rand.Reader)
	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			attrsMap := map[string]string{}
			for a := 0; a < count; a++ {
				attrsMap[fmt.Sprintf("attribute_%d", a)] = "true"
			}

			gamma := tkn20.Attributes{}
			gamma.FromMap(attrsMap)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				keys, _ := secKey.KeyGen(rand.Reader, gamma)
				runtime.KeepAlive(keys)
			}
		})
	}
}

func BenchmarkEncryptionAND(b *testing.B) {
	pubKey, _, _ := tkn20.Setup(rand.Reader)

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policyBuffer bytes.Buffer
			for p := 0; p < count-1; p++ {
				policyBuffer.WriteString("(attribute_" + strconv.Itoa(p) + ": true) and ")
			}
			policyBuffer.WriteString("(attribute_" + strconv.Itoa(count-1) + ": true)")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				policy := tkn20.Policy{}
				_ = policy.FromString(policyBuffer.String())
				cipher, _ := pubKey.Encrypt(rand.Reader, policy, []byte(msg))
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func BenchmarkEncryptionOR(b *testing.B) {
	pubKey, _, _ := tkn20.Setup(rand.Reader)

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policyBuffer bytes.Buffer
			for p := 0; p < count-1; p++ {
				policyBuffer.WriteString("(attribute_" + strconv.Itoa(p) + ": true) or ")
			}
			policyBuffer.WriteString("(attribute_" + strconv.Itoa(count-1) + ": true)")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				policy := tkn20.Policy{}
				_ = policy.FromString("(" + policyBuffer.String() + ")")
				cipher, _ := pubKey.Encrypt(rand.Reader, policy, []byte(msg))
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func TestCipherSizeAND(t *testing.T) {
	pubKey, _, _ := tkn20.Setup(rand.Reader)

	for size := range 25 {
		content := make([]byte, 1<<size)
		rand.Read(content)

		for _, count := range attributeCounts {
			var policyBuffer bytes.Buffer
			for p := 0; p < count-1; p++ {
				policyBuffer.WriteString("(attribute_" + strconv.Itoa(p) + ": true) and ")
			}
			policyBuffer.WriteString("(attribute_" + strconv.Itoa(count-1) + ": true)")
			policy := tkn20.Policy{}
			_ = policy.FromString(policyBuffer.String())
			cipher, _ := pubKey.Encrypt(rand.Reader, policy, []byte(msg))
			aes_cipher, _ := EncryptAES([]byte(msg), string(content))
			pure_abe, _ := pubKey.Encrypt(rand.Reader, policy, content)

			_ = cipher
			updateCSV("circl_tkn20_ct.csv", fmt.Sprint(count), "single "+fmt.Sprint(1<<size), fmt.Sprint(len(pure_abe)))
			updateCSV("circl_tkn20_ct.csv", fmt.Sprint(count), "hybrid "+fmt.Sprint(1<<size), fmt.Sprint(len(aes_cipher)+len(cipher)))
		}
	}
}

func TestCipherSizeOR(t *testing.T) {
	pubKey, _, _ := tkn20.Setup(rand.Reader)

	for size := range 25 {
		content := make([]byte, 1<<size)
		rand.Read(content)

		for _, count := range attributeCounts {
			var policyBuffer bytes.Buffer
			for p := 0; p < count-1; p++ {
				policyBuffer.WriteString("(attribute_" + strconv.Itoa(p) + ": true) or ")
			}
			policyBuffer.WriteString("(attribute_" + strconv.Itoa(count-1) + ": true)")
			policy := tkn20.Policy{}
			_ = policy.FromString(policyBuffer.String())
			cipher, _ := pubKey.Encrypt(rand.Reader, policy, []byte(msg))
			aes_cipher, _ := EncryptAES([]byte(msg), string(content))
			pure_abe, _ := pubKey.Encrypt(rand.Reader, policy, content)

			_ = cipher
			updateCSV("ciphertext_size_OR.csv", fmt.Sprint(count), "single "+fmt.Sprint(1<<size), fmt.Sprint(len(pure_abe)))
			updateCSV("ciphertext_size_OR.csv", fmt.Sprint(count), "hybrid "+fmt.Sprint(1<<size), fmt.Sprint(len(aes_cipher)+len(cipher)))
		}
	}
}

func BenchmarkDecryptionOR(b *testing.B) {
	pubKey, secKey, _ := tkn20.Setup(rand.Reader)

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policyBuffer bytes.Buffer
			for p := 0; p < count-1; p++ {
				policyBuffer.WriteString("(attribute_" + strconv.Itoa(p) + ": true) or ")
			}
			policyBuffer.WriteString("(attribute_" + strconv.Itoa(count-1) + ": true)")
			policy := tkn20.Policy{}
			_ = policy.FromString(policyBuffer.String())
			cipher, _ := pubKey.Encrypt(rand.Reader, policy, []byte(msg))

			gamma := tkn20.Attributes{}
			gamma.FromMap(map[string]string{fmt.Sprintf("attribute_%d", count-1): "true"})

			keys, _ := secKey.KeyGen(rand.Reader, gamma)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				text, e := keys.Decrypt(cipher)
				if e != nil {
					panic(e)
				}
				runtime.KeepAlive(text)
			}
		})
	}
}

func BenchmarkDecryptionAND(b *testing.B) {
	pubKey, secKey, _ := tkn20.Setup(rand.Reader)

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policyBuffer bytes.Buffer
			for p := 0; p < count-1; p++ {
				policyBuffer.WriteString("(attribute_" + strconv.Itoa(p) + ": true) and ")
			}
			policyBuffer.WriteString("(attribute_" + strconv.Itoa(count-1) + ": true)")
			policy := tkn20.Policy{}
			_ = policy.FromString(policyBuffer.String())
			cipher, _ := pubKey.Encrypt(rand.Reader, policy, []byte(msg))

			attrsMap := map[string]string{}
			for a := 0; a < count; a++ {
				attrsMap[fmt.Sprintf("attribute_%d", a)] = "true"
			}

			gamma := tkn20.Attributes{}
			gamma.FromMap(attrsMap)

			keys, _ := secKey.KeyGen(rand.Reader, gamma)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				text, e := keys.Decrypt(cipher)
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
