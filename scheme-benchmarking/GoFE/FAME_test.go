package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/fentec-project/gofe/abe"
)

func TestEndToEnd(t *testing.T) {

}

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

	attributeCounts := []int{1, 5, 10, 20, 50, 100}

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

	attributeCounts := []int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50}
	msg := "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul"

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " AND ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {

				cipher, _ := a.Encrypt(msg, msp, pubKey)
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func BenchmarkEncryptionOR(b *testing.B) {
	a := abe.NewFAME()
	pubKey, _, _ := a.GenerateMasterKeys()

	attributeCounts := []int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50}
	msg := "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul"

	for _, count := range attributeCounts {
		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			var policy bytes.Buffer
			for p := 0; p < count-1; p++ {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " OR ")
			}
			policy.WriteString("attribute_" + strconv.Itoa(count-1))
			msp, _ := abe.BooleanToMSP(policy.String(), false)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {

				cipher, _ := a.Encrypt(msg, msp, pubKey)
				runtime.KeepAlive(cipher)
			}
		})
	}
}

func BenchmarkDecryption(b *testing.B) {

}

func main() {
	res := testing.Benchmark(BenchmarkEncryptionOR)
	fmt.Printf("%s\n%#[1]v\n", res)
}
