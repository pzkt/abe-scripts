package crypto

import (
	"github.com/fentec-project/gofe/abe"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

type ABEscheme struct {
	Scheme    *abe.FAME
	PublicKey *abe.FAMEPubKey
	SecretKey *abe.FAMESecKey
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

func (s *ABEscheme) KeyGen(attributes []string) []byte {
	return utils.ToBytes(utils.Assure(s.Scheme.GenerateAttribKeys(attributes, s.SecretKey)))
}

func (s *ABEscheme) Encrypt(data []byte, policy string) []byte {
	msp, _ := abe.BooleanToMSP(policy, false)
	cipher, _ := s.Scheme.Encrypt(string(data), msp, s.PublicKey)
	return utils.ToBytes(cipher)
}

func Decrypt(ciphertext []byte, secret_key []byte) []byte {
	return []byte{}
}
