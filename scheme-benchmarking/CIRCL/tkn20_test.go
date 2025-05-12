package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/cloudflare/circl/abe/cpabe/tkn20"
)

var attributeCounts = [...]int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50}
var msg = "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul"

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
