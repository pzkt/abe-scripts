package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
)

type patient struct {
	ID               string           `json:"id"`
	Name             name             `json:"name"`
	DOB              time.Time        `json:"date_of_birth"`
	Address          string           `json:"address"`
	Phone            string           `json:"phone"`
	Email            string           `json:"email"`
	Insurance        string           `json:"insurance"`
	EmergencyContact emergencyContact `json:"emergency_contact"`
	CreatedAt        time.Time        `json:"created_date"`
	Records          []any            `json:"records"`
}

type name struct {
	NamePrefix string `json:"name_prefix"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
}

type emergencyContact struct {
	Name    name   `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type BaseMedicalRecord struct {
	PatientID  string    `json:"patient_id"`
	ProviderID string    `json:"provider_id"`
	RecordDate time.Time `json:"date"`
	Notes      string    `json:"notes"`
}

type CardiologyRecord struct {
	BaseMedicalRecord
	BloodPressure      int    `json:"blood_pressure"`
	HeartRate          int    `json:"heart_rate"`
	StressTestResults  string `json:"stress_test_results"`
	CardiacMedications string `json:"cardiac_medications"`
	EFPercentage       int    `json:"ef_percentage"`
}

type DermatologyRecord struct {
	BaseMedicalRecord
	SkinType          string `json:"skin_type"`
	LesionLocation    string `json:"lesion_location"`
	LesionDescription string `json:"lesion_description"`
	TreatmentPlan     string `json:"treatment_plan"`
	UVExposureHistory string `json:"uv_exposure_history"`
}

type HematologyRecord struct {
	BaseMedicalRecord
	Hemoglobin          float64 `json:"hemoglobin"`
	Hematocrit          int     `json:"hematocrit"`
	WhiteBloodCellCount int     `json:"wbc_count"`
	PlateletCount       int     `json:"platelet_count"`
	BloodSmearFindings  string  `json:"blood_smear_findings,omitempty"`
	BleedingTendency    bool    `json:"bleeding_tendency,omitempty"`
}

type NeurologyRecord struct {
	BaseMedicalRecord
	MentalStatus    string `json:"mental_status"`
	CranialNerves   string `json:"cranial_nerves"`
	MotorFunction   string `json:"motor_function"`
	SensoryFunction string `json:"sensory_function"`
	Reflexes        string `json:"reflexes"`
	Coordination    string `json:"coordination"`
	ImagingResults  string `json:"imaging_results,omitempty"`
}

type OncologyRecord struct {
	BaseMedicalRecord
	CancerType        string    `json:"cancer_type"`
	CancerStage       string    `json:"cancer_stage,omitempty"`
	TumorLocation     string    `json:"tumor_location"`
	Biomarkers        string    `json:"biomarkers,omitempty"`
	TreatmentPlan     string    `json:"treatment_plan"`
	LastTreatmentDate time.Time `json:"last_treatment_date,omitempty"`
}

var insuranceProviders = [...]string{"Aetna", "Blue Cross Blue Shield", "Cigna", "Humana", "Medicare", "Medicaid"}
var SkinType = [...]string{"I", "II", "III", "IV", "V", "VI"}
var CancerTypes = []string{"Adenocarcinoma", "Osteosarcoma", "Chondrosarcoma", "Glioblastoma", "Astrocytoma", "Pancreatic ductal adenocarcinoma", "Thyroid papillary carcinoma", "Melanoma"}

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

	var patients []patient
	for i := 0; i < count; i++ {
		new_patient := generatePatient()

		for j := 0; j < fake.Number(0, 6); j++ {
			new_patient.Records = append(new_patient.Records, generateRandomRecord(new_patient.ID))
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

func generateName() name {
	namePrefix := ""
	if fake.Bool() {
		namePrefix = fake.NamePrefix()
	}

	return name{
		NamePrefix: namePrefix,
		FirstName:  fake.FirstName(),
		LastName:   fake.LastName(),
	}
}

func generateEmergencyContact() emergencyContact {
	return emergencyContact{
		Name:    generateName(),
		Phone:   fake.Phone(),
		Email:   fake.Email(),
		Address: fake.Address().Address,
	}
}

func generatePatient() patient {
	return patient{
		ID:               fake.DigitN(8),
		Name:             generateName(),
		DOB:              fake.Date(),
		Address:          fake.Address().Address,
		Phone:            fake.Phone(),
		Email:            fake.Email(),
		Insurance:        insuranceProviders[rand.Intn(len(insuranceProviders))],
		EmergencyContact: generateEmergencyContact(),
		CreatedAt:        fake.Date(),
		Records:          []any{},
	}
}

func generateRandomRecord(patient_id string) any {
	switch fake.Number(0, 4) {
	case 0:
		return generateCardiologyRecord(patient_id)
	case 1:
		return generateDermatologyRecord(patient_id)
	case 2:
		return generateHematologyRecord(patient_id)
	case 3:
		return generateNeurologyRecord(patient_id)
	default:
		return generateOncologyRecord(patient_id)
	}
}

func generateBaseMedicalRecord(patient_id string) BaseMedicalRecord {
	return BaseMedicalRecord{
		PatientID:  patient_id,
		ProviderID: fake.DigitN(6),
		RecordDate: fake.Date(),
		Notes:      fake.LoremIpsumSentence(fake.Number(2, 15)),
	}
}

func generateCardiologyRecord(patient_id string) CardiologyRecord {
	return CardiologyRecord{
		BaseMedicalRecord:  generateBaseMedicalRecord(patient_id),
		BloodPressure:      fake.Number(100, 200),
		HeartRate:          fake.Number(30, 200),
		StressTestResults:  fake.LoremIpsumSentence(fake.Number(2, 15)),
		CardiacMedications: fake.LoremIpsumSentence(fake.Number(2, 15)),
		EFPercentage:       fake.Number(0, 100),
	}
}

func generateDermatologyRecord(patient_id string) DermatologyRecord {
	return DermatologyRecord{
		BaseMedicalRecord: generateBaseMedicalRecord(patient_id),
		SkinType:          SkinType[rand.Intn(len(SkinType))],
		LesionLocation:    fake.LoremIpsumWord(),
		LesionDescription: fake.LoremIpsumSentence(fake.Number(2, 15)),
		TreatmentPlan:     fake.LoremIpsumSentence(fake.Number(2, 15)),
		UVExposureHistory: fake.LoremIpsumSentence(fake.Number(2, 15)),
	}
}

func generateHematologyRecord(patient_id string) HematologyRecord {
	return HematologyRecord{
		BaseMedicalRecord:   generateBaseMedicalRecord(patient_id),
		Hemoglobin:          fake.Float64Range(10.5, 20.5),
		Hematocrit:          fake.Number(0, 100),
		WhiteBloodCellCount: fake.Number(4000, 11000),
		PlateletCount:       fake.Number(150000, 450000),
		BloodSmearFindings:  fake.LoremIpsumSentence(fake.Number(2, 15)),
		BleedingTendency:    fake.Bool(),
	}
}

func generateNeurologyRecord(patient_id string) NeurologyRecord {
	return NeurologyRecord{
		BaseMedicalRecord: generateBaseMedicalRecord(patient_id),
		MentalStatus:      fake.LoremIpsumSentence(fake.Number(2, 5)),
		CranialNerves:     fake.LoremIpsumSentence(fake.Number(2, 5)),
		MotorFunction:     fake.LoremIpsumSentence(fake.Number(2, 5)),
		SensoryFunction:   fake.LoremIpsumSentence(fake.Number(2, 5)),
		Reflexes:          fake.LoremIpsumSentence(fake.Number(2, 5)),
		Coordination:      fake.LoremIpsumSentence(fake.Number(2, 5)),
		ImagingResults:    fake.LoremIpsumSentence(fake.Number(2, 5)),
	}
}

func generateOncologyRecord(patient_id string) OncologyRecord {
	return OncologyRecord{
		BaseMedicalRecord: generateBaseMedicalRecord(patient_id),
		CancerType:        CancerTypes[rand.Intn(len(CancerTypes))],
		TumorLocation:     fake.LoremIpsumSentence(fake.Number(2, 5)),
		Biomarkers:        fake.LoremIpsumSentence(fake.Number(2, 5)),
		TreatmentPlan:     fake.LoremIpsumSentence(fake.Number(10, 25)),
		LastTreatmentDate: fake.Date(),
	}
}
