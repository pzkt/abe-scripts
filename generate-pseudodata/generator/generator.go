package generator

import (
	"fmt"
	"math/rand"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
)

type Patient struct {
	ID               string           `json:"id"`
	Name             Name             `json:"name"`
	DOB              time.Time        `json:"date_of_birth"`
	Address          string           `json:"address"`
	Phone            string           `json:"phone"`
	Email            string           `json:"email"`
	Insurance        string           `json:"insurance"`
	EmergencyContact EmergencyContact `json:"emergency_contact"`
	CreatedAt        time.Time        `json:"created_date"`
	Records          []any            `json:"records"`
}

type Name struct {
	NamePrefix string `json:"name_prefix"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
}

type EmergencyContact struct {
	Name    Name   `json:"name"`
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

func Test() {
	fmt.Println("test")
}

func GenerateName() Name {
	namePrefix := ""
	if fake.Bool() {
		namePrefix = fake.NamePrefix()
	}

	return Name{
		NamePrefix: namePrefix,
		FirstName:  fake.FirstName(),
		LastName:   fake.LastName(),
	}
}

func GenerateEmergencyContact() EmergencyContact {
	return EmergencyContact{
		Name:    GenerateName(),
		Phone:   fake.Phone(),
		Email:   fake.Email(),
		Address: fake.Address().Address,
	}
}

func GeneratePatient() Patient {
	return Patient{
		ID:               fake.DigitN(8),
		Name:             GenerateName(),
		DOB:              fake.Date(),
		Address:          fake.Address().Address,
		Phone:            fake.Phone(),
		Email:            fake.Email(),
		Insurance:        insuranceProviders[rand.Intn(len(insuranceProviders))],
		EmergencyContact: GenerateEmergencyContact(),
		CreatedAt:        fake.Date(),
		Records:          []any{},
	}
}

func GenerateRandomRecord(patient_id string) any {
	switch fake.Number(0, 4) {
	case 0:
		return GenerateCardiologyRecord(patient_id)
	case 1:
		return GenerateDermatologyRecord(patient_id)
	case 2:
		return GenerateHematologyRecord(patient_id)
	case 3:
		return GenerateNeurologyRecord(patient_id)
	default:
		return GenerateOncologyRecord(patient_id)
	}
}

func GenerateBaseMedicalRecord(patient_id string) BaseMedicalRecord {
	return BaseMedicalRecord{
		PatientID:  patient_id,
		ProviderID: fake.DigitN(6),
		RecordDate: fake.Date(),
		Notes:      fake.LoremIpsumSentence(fake.Number(2, 15)),
	}
}

func GenerateCardiologyRecord(patient_id string) CardiologyRecord {
	return CardiologyRecord{
		BaseMedicalRecord:  GenerateBaseMedicalRecord(patient_id),
		BloodPressure:      fake.Number(100, 200),
		HeartRate:          fake.Number(30, 200),
		StressTestResults:  fake.LoremIpsumSentence(fake.Number(2, 15)),
		CardiacMedications: fake.LoremIpsumSentence(fake.Number(2, 15)),
		EFPercentage:       fake.Number(0, 100),
	}
}

func GenerateDermatologyRecord(patient_id string) DermatologyRecord {
	return DermatologyRecord{
		BaseMedicalRecord: GenerateBaseMedicalRecord(patient_id),
		SkinType:          SkinType[rand.Intn(len(SkinType))],
		LesionLocation:    fake.LoremIpsumWord(),
		LesionDescription: fake.LoremIpsumSentence(fake.Number(2, 15)),
		TreatmentPlan:     fake.LoremIpsumSentence(fake.Number(2, 15)),
		UVExposureHistory: fake.LoremIpsumSentence(fake.Number(2, 15)),
	}
}

func GenerateHematologyRecord(patient_id string) HematologyRecord {
	return HematologyRecord{
		BaseMedicalRecord:   GenerateBaseMedicalRecord(patient_id),
		Hemoglobin:          fake.Float64Range(10.5, 20.5),
		Hematocrit:          fake.Number(0, 100),
		WhiteBloodCellCount: fake.Number(4000, 11000),
		PlateletCount:       fake.Number(150000, 450000),
		BloodSmearFindings:  fake.LoremIpsumSentence(fake.Number(2, 15)),
		BleedingTendency:    fake.Bool(),
	}
}

func GenerateNeurologyRecord(patient_id string) NeurologyRecord {
	return NeurologyRecord{
		BaseMedicalRecord: GenerateBaseMedicalRecord(patient_id),
		MentalStatus:      fake.LoremIpsumSentence(fake.Number(2, 5)),
		CranialNerves:     fake.LoremIpsumSentence(fake.Number(2, 5)),
		MotorFunction:     fake.LoremIpsumSentence(fake.Number(2, 5)),
		SensoryFunction:   fake.LoremIpsumSentence(fake.Number(2, 5)),
		Reflexes:          fake.LoremIpsumSentence(fake.Number(2, 5)),
		Coordination:      fake.LoremIpsumSentence(fake.Number(2, 5)),
		ImagingResults:    fake.LoremIpsumSentence(fake.Number(2, 5)),
	}
}

func GenerateOncologyRecord(patient_id string) OncologyRecord {
	return OncologyRecord{
		BaseMedicalRecord: GenerateBaseMedicalRecord(patient_id),
		CancerType:        CancerTypes[rand.Intn(len(CancerTypes))],
		TumorLocation:     fake.LoremIpsumSentence(fake.Number(2, 5)),
		Biomarkers:        fake.LoremIpsumSentence(fake.Number(2, 5)),
		TreatmentPlan:     fake.LoremIpsumSentence(fake.Number(10, 25)),
		LastTreatmentDate: fake.Date(),
	}
}
