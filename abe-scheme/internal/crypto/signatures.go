/*

Functions for signing

*/

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"math/big"

	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

func GenerateSignatureKey() *ecdsa.PrivateKey {
	return utils.Assure(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))
}

func Sign(privateKey *ecdsa.PrivateKey, data []byte) []byte {
	hashed := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed[:])

	//crash and burn on error
	if err != nil {
		log.Fatal(err)
	}

	sig := append(r.Bytes(), s.Bytes()...)
	return sig
}

func Verify(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	hashed := sha256.Sum256(data)

	r := new(big.Int).SetBytes(signature[:len(signature)/2])
	s := new(big.Int).SetBytes(signature[len(signature)/2:])

	return ecdsa.Verify(publicKey, hashed[:], r, s)
}
