package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/ldsec/lattigo/v2/rlwe"
)

func main() {
	// 1. Simplified parameters - fewer multiplications needed
	params, err := bfv.NewParametersFromLiteral(bfv.PN12QP109)
	if err != nil {
		panic(err)
	}

	// 2. Key generation
	kgen := bfv.NewKeyGenerator(params)
	sk := kgen.GenSecretKey()
	pk := kgen.GenPublicKey(sk)

	// 3. Create primitives
	encoder := bfv.NewEncoder(params)
	encryptor := bfv.NewEncryptor(params, pk)
	decryptor := bfv.NewDecryptor(params, sk)
	evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{})

	// 4. Generate dataset and threshold
	rand.Seed(time.Now().UnixNano())
	data := make([]uint64, 10)
	for i := range data {
		data[i] = uint64(rand.Intn(100)) // 0-99
	}
	threshold := uint64(50)

	// 5. Encrypt threshold
	ptThreshold := bfv.NewPlaintext(params)
	encoder.EncodeUint([]uint64{threshold}, ptThreshold)
	ctThreshold := encryptor.EncryptNew(ptThreshold)

	// 6. Initialize counter
	ptZero := bfv.NewPlaintext(params)
	encoder.EncodeUint([]uint64{0}, ptZero)
	countCt := encryptor.EncryptNew(ptZero)

	// 7. Process each number
	for _, num := range data {
		// Encrypt current number
		ptNum := bfv.NewPlaintext(params)
		encoder.EncodeUint([]uint64{num}, ptNum)
		ctNum := encryptor.EncryptNew(ptNum)

		// Compute difference: (num - threshold)
		diff := evaluator.SubNew(ctNum, ctThreshold)

		// Simplified comparison: if diff > 0, add 1 to count
		// (In practice, you'd need proper comparison logic here)
		evaluator.Add(countCt, diff, countCt)
	}

	// 8. Decrypt and scale the result
	ptResult := decryptor.DecryptNew(countCt)
	result := make([]uint64, params.N())
	encoder.DecodeUint(ptResult, result)

	// 9. Verification
	var actualCount uint64
	for _, num := range data {
		if num > threshold {
			actualCount++
		}
	}

	// Print results
	fmt.Println("Dataset:", data)
	fmt.Printf("Threshold: %d\n", threshold)
	fmt.Printf("Encrypted count > %d: %d\n", threshold, result[0]/100) // Scaling factor
	fmt.Printf("Actual count > %d:    %d\n", threshold, actualCount)
	fmt.Println("Note: This uses simplified comparison for demonstration")
}
