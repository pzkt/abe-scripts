package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/pzkt/abe-scripts/generate-pseudodata/generator"
)

var attributeCounts = [...]int{1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50}

// average file size: 443 bytes
func BenchmarkUploadSmallEntry(b *testing.B) {
	for n := 0; n < b.N; n++ {
		env := setup()
		record := generator.GenerateRandomRecord(uuid.NewString())
		env.addEntry("table_one", record, "Radiology AND Masked-Research", "Radiology AND Masked-Research")
	}
}

// average file size: 42.35 kilobytes
func BenchmarkUploadMediumEntry(b *testing.B) {
	for n := 0; n < b.N; n++ {
		env := setup()

		new_patient := generator.GeneratePatient()

		for j := 0; j < 100; j++ {
			new_patient.Records = append(new_patient.Records, generator.GenerateRandomRecord(new_patient.ID))
		}

		env.addEntry("table_one", new_patient, "Radiology AND Masked-Research", "Radiology AND Masked-Research")
	}
}

// average file size: 39.83 megabytes
func BenchmarkUploadLargeEntry(b *testing.B) {
	for n := 0; n < b.N; n++ {
		env := setup()

		new_patient := generator.GeneratePatient()

		for j := 0; j < 100000; j++ {
			new_patient.Records = append(new_patient.Records, generator.GenerateRandomRecord(new_patient.ID))
		}
		//generating the data takes a considerable amount of time. Don't count it to the total
		b.ResetTimer()

		env.addEntry("table_one", new_patient, "Radiology AND Masked-Research", "Radiology AND Masked-Research")
	}
}

func BenchmarkModifyEntry(b *testing.B) {
	env := setup()

	new_patient := generator.GeneratePatient()

	for j := 0; j < 100; j++ {
		new_patient.Records = append(new_patient.Records, generator.GenerateRandomRecord(new_patient.ID))
	}
	entryUUID := env.addEntry("table_one", new_patient, "Radiology AND Masked-Research", "Radiology AND Masked-Research")

	for n := 0; n < b.N; n++ {
		new_patient = generator.GeneratePatient()

		for j := 0; j < 100; j++ {
			new_patient.Records = append(new_patient.Records, generator.GenerateRandomRecord(new_patient.ID))
		}
		env.modifyEntry("table_one", new_patient, "Radiology AND Masked-Research", "Radiology AND Masked-Research", entryUUID)
	}
}

func BenchmarkGetEntry(b *testing.B) {
	env := setup()

	new_patient := generator.GeneratePatient()

	for j := 0; j < 100; j++ {
		new_patient.Records = append(new_patient.Records, generator.GenerateRandomRecord(new_patient.ID))
	}

	entryUUID := env.addEntry("table_one", new_patient, "Radiology AND Masked-Research", "Radiology AND Masked-Research")

	key := requestNewKey([]string{"General-Purpose"})

	for n := 0; n < b.N; n++ {
		data := env.getEntry("table_one", entryUUID).Data
		decrypted_data := env.abeScheme.Decrypt(data, key)
		runtime.KeepAlive(decrypted_data)
	}
}

func BenchmarkUploadVariablePolicy(b *testing.B) {
	env := setup()

	//small: 443 medium: 42350 large: 39830000
	content := make([]byte, 39830000)
	rand.Read(content)

	for _, count := range attributeCounts {

		gamma := make([]string, count)
		for a := 0; a < count; a++ {
			gamma[a] = fmt.Sprintf("attribute_%d", a)
		}

		key := requestNewKey(gamma)

		var policy bytes.Buffer
		for p := 0; p < count-1; p++ {
			if p%2 == 0 {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " OR ")
			} else {
				policy.WriteString("attribute_" + strconv.Itoa(p) + " AND ")
			}

		}
		policy.WriteString("attribute_" + strconv.Itoa(count-1))

		entryUUID := env.addEntry("table_one", content, policy.String(), policy.String())

		b.Run(fmt.Sprintf("Attributes_%d", count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				data := env.getEntry("table_one", entryUUID).Data
				decrypted_data := env.abeScheme.Decrypt(data, key)
				runtime.KeepAlive(decrypted_data)
			}
		})
	}
}
