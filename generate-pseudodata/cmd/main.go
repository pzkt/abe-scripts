package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/pzkt/abe-scripts/generate-pseudodata/generator"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Incorrect args: provide one argument for the number of generated patient records")
		return
	}

	count, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Incorrect args: must be a number")
		return
	}

	var patients []generator.Patient
	for i := 0; i < count; i++ {
		new_patient := generator.GeneratePatient()

		for j := 0; j < fake.Number(0, 6); j++ {
			new_patient.Records = append(new_patient.Records, generator.GenerateRandomRecord(new_patient.ID))
		}

		patients = append(patients, new_patient)
	}
	writeJSON("patient_data.json", patients)
}

func writeJSON(filename string, data interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "")
	if err := encoder.Encode(data); err != nil {
		panic(err)
	}
}
