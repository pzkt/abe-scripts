package crypto

import (
	"github.com/fentec-project/gofe/abe"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

type ABEscheme struct {
	scheme    *abe.FAME
	PublicKey *abe.FAMEPubKey
}

func Setup() *ABEscheme {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()
	return &ABEscheme{
		scheme:    a,
		PublicKey: pubKey,
	}
}

func (s *ABEscheme) Encrypt(data []byte, policy string) []byte {
	msp, _ := abe.BooleanToMSP(policy, false)
	cipher, _ := s.scheme.Encrypt(string(data), msp, s.PublicKey)
	return utils.ToBytes(cipher)
}

/* func Decrypt(ciphertext []byte, secret_key []byte) []byte {

}
*/
