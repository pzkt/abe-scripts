package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/ldsec/lattigo/v2/rlwe"
)

func main() {
	// 1. Parameters setup (128-bit security)
	params, err := bfv.NewParametersFromLiteral(bfv.PN12QP109)
	if err != nil {
		panic(err)
	}

	// 2. Key generation
	kgen := bfv.NewKeyGenerator(params)
	sk := kgen.GenSecretKey()
	pk := kgen.GenPublicKey(sk)
	rlk := kgen.GenRelinearizationKey(sk, 1)

	// 3. Create primitives
	encoder := bfv.NewEncoder(params)
	encryptor := bfv.NewEncryptor(params, pk)
	decryptor := bfv.NewDecryptor(params, sk)
	evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{Rlk: rlk})

	// 4. Generate 50 random numbers (0-99)
	rand.Seed(time.Now().UnixNano())
	data := make([]uint64, 1000)
	for i := range data {
		data[i] = uint64(rand.Intn(100))
	}

	// 5. Encrypt each number individually (simple approach)
	ciphertexts := make([]*bfv.Ciphertext, len(data))
	for i, num := range data {
		pt := bfv.NewPlaintext(params)
		encoder.EncodeUint([]uint64{num}, pt)
		ciphertexts[i] = encryptor.EncryptNew(pt)
	}

	// 6. Homomorphic summation
	sumCt := ciphertexts[0].CopyNew()
	for _, ct := range ciphertexts[1:] {
		evaluator.Add(sumCt, ct, sumCt) // sumCt += ct

	}

	// 7. Decrypt
	ptResult := decryptor.DecryptNew(sumCt)
	result := make([]uint64, params.N())
	encoder.DecodeUint(ptResult, result)

	// 8. Verification
	var actualSum uint64
	for _, v := range data {
		actualSum += v
	}

	fmt.Println("=== BFV Summation Results ===")
	fmt.Printf("Homomorphic sum: %d\n", result[0])
	fmt.Printf("Actual sum:      %d\n", actualSum)
	fmt.Printf("Match:           %v\n", result[0] == actualSum)
	fmt.Printf("Number of adds:  %d\n", len(data)-1)
}
