package crypto

import (
	"crypto/rand"

	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

func generateSymKey() []byte {
	key := make([]byte, 32) // AES-256
	utils.Assure(rand.Read(key))
	return key
}
