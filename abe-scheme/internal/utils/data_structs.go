package utils

type CardiologyRecord struct {
	Notes              string `json:"notes"`
	BloodPressure      int    `json:"blood_pressure"`
	HeartRate          int    `json:"heart_rate"`
	StressTestResults  string `json:"stress_test_results"`
	CardiacMedications string `json:"cardiac_medications"`
	EFPercentage       int    `json:"ef_percentage"`
}

type TestRecord struct {
	Notes string `json:"notes"`
	Goats int    `json:"goats"`
}
