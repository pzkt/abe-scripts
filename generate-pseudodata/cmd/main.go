package main

import (
	"fmt"
	"os"
	"strconv"

	fake "github.com/brianvoe/gofakeit/v7"
	_ "github.com/pzkt/abe-scripts/generate-pseudodata/cmd"
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

	var patients []Patient
	for i := 0; i < count; i++ {
		new_patient := GeneratePatient()

		for j := 0; j < fake.Number(0, 6); j++ {
			new_patient.Records = append(new_patient.Records, GenerateRandomRecord(new_patient.ID))
		}

		patients = append(patients, new_patient)
	}
	writeJSON("patient_data.json", patients)
}
