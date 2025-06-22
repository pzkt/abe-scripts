package main

import (
	"fmt"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/ldsec/lattigo/v2/rlwe"
)

func main() {
	// Step 1: Set up BFV parameters
	params, err := bfv.NewParametersFromLiteral(bfv.PN12QP109)
	if err != nil {
		panic(err)
	}

	// Step 2: Generate keys
	kgen := bfv.NewKeyGenerator(params)
	sk := kgen.GenSecretKey()
	pk := kgen.GenPublicKey(sk)
	rlk := kgen.GenRelinearizationKey(sk, 1)

	// Step 3: Create encryptor, evaluator, and decryptor
	encryptor := bfv.NewEncryptor(params, pk)
	decryptor := bfv.NewDecryptor(params, sk)
	evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{Rlk: rlk})

	// Step 4: Prepare plaintext values
	x := uint64(1344)
	y := uint64(5)

	// Encode plaintexts
	encoder := bfv.NewEncoder(params)
	ptX := bfv.NewPlaintext(params)
	ptY := bfv.NewPlaintext(params)
	encoder.EncodeUint([]uint64{x}, ptX)
	encoder.EncodeUint([]uint64{y}, ptY)

	// Step 5: Encrypt plaintexts
	ctX := encryptor.EncryptNew(ptX)
	ctY := encryptor.EncryptNew(ptY)

	fmt.Println(*ctX.Ciphertext.Value[0])

	// Step 6: Homomorphic operations
	// Addition: ctAdd = Enc(x + y)
	ctAdd := evaluator.AddNew(ctX, ctY)

	// Multiplication: ctMul = Enc(x * y)
	ctMul := evaluator.MulNew(ctX, ctY)

	// Step 7: Decrypt results
	ptAdd := bfv.NewPlaintext(params)
	ptMul := bfv.NewPlaintext(params)
	decryptor.Decrypt(ctAdd, ptAdd)
	decryptor.Decrypt(ctMul, ptMul)

	// Decode results
	resAdd := make([]uint64, params.N())
	encoder.DecodeUint(ptAdd, resAdd)

	//resAdd := encoder.DecodeUint(ptAdd)
	resMul := make([]uint64, params.N())
	encoder.DecodeUint(ptMul, resMul)

	//resMul := encoder.DecodeUint(ptMul)

	// Step 8: Print results
	fmt.Printf("Encrypted Addition: %d + %d = %d\n", x, y, resAdd[0])
	fmt.Printf("Encrypted Multiplication: %d * %d = %d\n", x, y, resMul[0])
}
