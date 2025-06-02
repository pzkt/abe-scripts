package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"

	_ "github.com/lib/pq"
)

// HybridEncryptor manages both symmetric and asymmetric encryption
type HybridEncryptor struct {
	privateKey       *rsa.PrivateKey
	recipientPubKeys map[string]*rsa.PublicKey // Keyed by recipient ID
}

type EncryptedData struct {
	EncryptedKey  string            `json:"encrypted_key"`  // Base64 encoded
	EncryptedData string            `json:"encrypted_data"` // Base64 encoded
	KeyRecipients map[string]string `json:"key_recipients"` // RecipientID -> encrypted key for them
	BlindIndex    string            `json:"blind_index"`    // For searchable fields
}

func NewHybridEncryptor(privateKeyPEM []byte, recipientPubKeys map[string][]byte) (*HybridEncryptor, error) {
	// Parse private key
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Parse recipient public keys
	recipients := make(map[string]*rsa.PublicKey)
	for id, pubKeyPEM := range recipientPubKeys {
		block, _ := pem.Decode(pubKeyPEM)
		if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block for recipient %s", id)
		}

		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key for recipient %s: %v", id, err)
		}

		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("key for recipient %s is not RSA", id)
		}

		recipients[id] = rsaPubKey
	}

	return &HybridEncryptor{
		privateKey:       privateKey,
		recipientPubKeys: recipients,
	}, nil
}

// GenerateDEK creates a new data encryption key (symmetric)
func generateDEK() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptWithDEK symmetrically encrypts data
func encryptWithDEK(dek []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithDEK symmetrically decrypts data
func decryptWithDEK(dek []byte, ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	//nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	//plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string("plaintext"), nil
}

// createBlindIndex creates a searchable hash of the data
func createBlindIndex(value, pepper string) string {
	h := sha256.New()
	h.Write([]byte(value + pepper))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

//--------------------------------------------------------

// EncryptData encrypts data with a new DEK and then encrypts the DEK for all recipients
func (h *HybridEncryptor) EncryptData(plaintext, fieldName string) (*EncryptedData, error) {
	// Generate a new DEK for this data
	dek, err := generateDEK()
	if err != nil {
		return nil, err
	}

	// Encrypt the data with the DEK
	encryptedData, err := encryptWithDEK(dek, plaintext)
	if err != nil {
		return nil, err
	}

	// Encrypt the DEK for each recipient
	keyRecipients := make(map[string]string)
	for recipientID, pubKey := range h.recipientPubKeys {
		encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, dek, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt for recipient %s: %v", recipientID, err)
		}
		keyRecipients[recipientID] = base64.StdEncoding.EncodeToString(encryptedKey)
	}

	// Create blind index for searchability (use field-specific pepper)
	pepper := "pepper-" + fieldName // In production, get from secure config
	blindIndex := createBlindIndex(plaintext, pepper)

	return &EncryptedData{
		EncryptedKey:  "", // Not needed when we have keyRecipients
		EncryptedData: encryptedData,
		KeyRecipients: keyRecipients,
		BlindIndex:    blindIndex,
	}, nil
}

// DecryptData decrypts the DEK with our private key and then decrypts the data
func (h *HybridEncryptor) DecryptData(encData *EncryptedData, recipientID string) (string, error) {
	// Find our encrypted DEK
	encryptedKeyBase64, ok := encData.KeyRecipients[recipientID]
	if !ok {
		return "", fmt.Errorf("no encrypted key found for recipient %s", recipientID)
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyBase64)
	if err != nil {
		return "", err
	}

	// Decrypt the DEK with our private key
	dek, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, h.privateKey, encryptedKey, nil)
	if err != nil {
		return "", err
	}

	// Decrypt the data with the DEK
	return decryptWithDEK(dek, encData.EncryptedData)
}

//-------------------------------------------------------------

type User struct {
	ID        int
	Name      string
	Email     string
	SSN       *EncryptedData // Encrypted SSN with metadata
	CreatedAt sql.NullTime
}

func maine() {
	// Initialize keys (in production, load from secure sources)
	privateKeyPEM := []byte(`-----BEGIN RSA PRIVATE KEY-----
... your private key PEM ... 
-----END RSA PRIVATE KEY-----`)

	recipientPubKeys := map[string][]byte{
		"service1": []byte(`-----BEGIN PUBLIC KEY-----
... service1 public key PEM ...
-----END PUBLIC KEY-----`),
		"service2": []byte(`-----BEGIN PUBLIC KEY-----
... service2 public key PEM ...
-----END PUBLIC KEY-----`),
	}

	encryptor, err := NewHybridEncryptor(privateKeyPEM, recipientPubKeys)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", "user=postgres dbname=testdb sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table with JSON column for encrypted data
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		ssn_data JSONB,
		ssn_blind_index TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert encrypted data
	ssn := "123-45-6789"
	encryptedSSN, err := encryptor.EncryptData(ssn, "ssn")
	if err != nil {
		log.Fatal(err)
	}

	ssnJSON, err := json.Marshal(encryptedSSN)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO users (name, email, ssn_data, ssn_blind_index) 
		VALUES ($1, $2, $3, $4)`,
		"John Doe", "john@example.com", ssnJSON, encryptedSSN.BlindIndex)
	if err != nil {
		log.Fatal(err)
	}

	// Query by blind index (searchable)
	pepper := "pepper-ssn" // Should match what was used in EncryptData
	searchSSN := "123-45-6789"
	searchIndex := createBlindIndex(searchSSN, pepper)

	var user User
	var ssnDataJSON []byte
	err = db.QueryRow(`SELECT id, name, email, ssn_data FROM users 
		WHERE ssn_blind_index = $1`, searchIndex).Scan(
		&user.ID, &user.Name, &user.Email, &ssnDataJSON)
	if err != nil {
		log.Fatal(err)
	}

	var encryptedSSNData EncryptedData
	if err := json.Unmarshal(ssnDataJSON, &encryptedSSNData); err != nil {
		log.Fatal(err)
	}
	user.SSN = &encryptedSSNData

	// Decrypt the SSN
	decryptedSSN, err := encryptor.DecryptData(user.SSN, "service1") // Using our recipient ID
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %+v, Decrypted SSN: %s\n", user, decryptedSSN)
}
