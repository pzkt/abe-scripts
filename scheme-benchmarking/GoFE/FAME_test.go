package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/fentec-project/gofe/abe"
)

var attributeCounts = [...]int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50}
var msg = "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul"

func BenchmarkSetup(b *testing.B) {
	for n := 0; n < b.N; n++ {
		a := abe.NewFAME()
		pubKey, secKey, _ := a.GenerateMasterKeys()
		runtime.KeepAlive(pubKey)
		runtime.KeepAlive(secKey)
	}
}

func BenchmarkKeyGen(b *testing.B) {
	a := abe.NewFAME()
	_, secKey, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			gamma := make([]string, count)
			for a := 0; a < count; a++ {
				gamma[a] = fmt.Sprintf("attribute_%d", a)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				keys, _ := a.GenerateAttribKeys(gamma, secKey)
				runtime.KeepAlive(keys)
			}
		})
	}
}

func BenchmarkEncryptionAND(b *testing.B) {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " AND ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				msp, _ := abe.BooleanToMSP(policy.String(), false)
				cipher, _ := a.Encrypt(msg, msp, pubKey)
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func BenchmarkEncryptionOR(b *testing.B) {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " OR ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				msp, _ := abe.BooleanToMSP(policy.String(), false)
				cipher, _ := a.Encrypt(msg, msp, pubKey)
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func BenchmarkDecryptionOR(b *testing.B) {
	a := abe.NewFAME()
	pubKey, secKey, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " OR ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)
			cipher, _ := a.Encrypt(msg, msp, pubKey)

			keys, _ := a.GenerateAttribKeys([]string{fmt.Sprintf("attribute_%d", count-1)}, secKey)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				text, e := a.Decrypt(cipher, keys, pubKey)
				if e != nil {
					panic(e)
				}
				runtime.KeepAlive(text)
			}
		})
	}
}

func BenchmarkDecryptionAND(b *testing.B) {
	a := abe.NewFAME()
	pubKey, secKey, _ := a.GenerateMasterKeys()

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " AND ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)
			cipher, _ := a.Encrypt(msg, msp, pubKey)

			gamma := make([]string, count)
			for a := 0; a < count; a++ {
				gamma[a] = fmt.Sprintf("attribute_%d", a)
			}
			keys, _ := a.GenerateAttribKeys(gamma, secKey)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				text, e := a.Decrypt(cipher, keys, pubKey)
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
