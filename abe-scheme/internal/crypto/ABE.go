package crypto

import (
	"fmt"

	"github.com/fentec-project/gofe/abe"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

type ABEscheme struct {
	Scheme    *abe.FAME
	PublicKey *abe.FAMEPubKey
	SecretKey *abe.FAMESecKey
}

type ABEpublicKey struct {
	PublicKey *abe.FAMEPubKey
}

func Setup() *ABEscheme {
	a := abe.NewFAME()
	pubKey, secKey, _ := a.GenerateMasterKeys()
	return &ABEscheme{
		Scheme:    a,
		PublicKey: pubKey,
		SecretKey: secKey,
	}
}

func (s *ABEscheme) EndToEndTest() {

	cipher := s.Encrypt(utils.ToBytes("wow schgloopy"), "test OR few")

	key := s.KeyGen([]string{"test", "wow"})

	text := s.Decrypt(cipher, key)

	fmt.Println(string(text))
}

func (s *ABEscheme) KeyGen(attributes []string) []byte {
	key := utils.Assure(s.Scheme.GenerateAttribKeys(attributes, s.SecretKey))
	keyBytes := utils.ToBytes(key)
	return keyBytes
}

func (s *ABEscheme) Encrypt(data []byte, policy string) []byte {
	msp, _ := abe.BooleanToMSP(policy, false)
	cipher, _ := s.Scheme.Encrypt(string(data), msp, s.PublicKey)

	return utils.ToBytes(cipher)
}

func (s *ABEscheme) Decrypt(ciphertext []byte, secret_key []byte) []byte {
	var cipher abe.FAMECipher
	utils.FromBytes(ciphertext, &cipher)

	var key abe.FAMEAttribKeys
	utils.FromBytes(secret_key, &key)

	plaintext := utils.Assure(s.Scheme.Decrypt(&cipher, &key, s.PublicKey))
	return utils.ToBytes(plaintext)
}
