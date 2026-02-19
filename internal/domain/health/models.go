package health

import "time"

type HealthMetric struct {
	ID         int64
	Timestamp  string
	Platform   string
	MetricType string
	Value      float64
	Unit       string
	SourceFile string
}

type WithingsActivity struct {
	Timestamp  string
	Platform   string
	MetricType string
	Value      float64
	Unit       string
	SourceFile string
}

type SpO2Record struct {
	Timestamp  string
	Platform   string
	MetricType string
	Value      float64
	Unit       string
	SourceFile string
}

type MFPWeightRecord struct {
	Timestamp  string
	Platform   string
	MetricType string
	Value      float64
	Unit       string
	SourceFile string
}

type ImportResult struct {
	TotalRows int
	Inserted  int
	Skipped   int
	Errors    []ImportError
	Duration  time.Duration
}

type ImportError struct {
	Row  int
	Col  int
	Err  error
	Data map[string]string
}

func (e ImportError) Error() string {
	return e.Err.Error()
}

func (r ImportResult) IsSuccess() bool {
	return len(r.Errors) == 0
}

const (
	PlatformWithings     = "withings"
	PlatformMyFitnessPal = "myfitnesspal"

	MetricTypeSteps      = "steps"
	MetricTypeCalories   = "calories"
	MetricTypeDistance   = "distance"
	MetricTypeElevation  = "elevation"
	MetricTypeHrAverage  = "hr_average"
	MetricTypeHrMin      = "hr_min"
	MetricTypeHrMax      = "hr_max"
	MetricTypeSpO2       = "spo2"
	MetricTypeWeight     = "weight"
	MetricTypeFatMass    = "fat_mass"
	MetricTypeBoneMass   = "bone_mass"
	MetricTypeMuscleMass = "muscle_mass"
	MetricTypeHydration  = "hydration"

	UnitCount   = "count"
	UnitKcal    = "kcal"
	UnitMeters  = "meters"
	UnitBpm     = "bpm"
	UnitPercent = "%"
	UnitKg      = "kg"
	UnitLb      = "lb"
	UnitSeconds = "seconds"
)

type WeightRecord struct {
	ID           int64
	Timestamp    string
	WeightLb     float64
	FatMassLb    float64
	BoneMassLb   float64
	MuscleMassLb float64
	HydrationLb  float64
	Comments     string
	Source       string
}

type SleepRecord struct {
	ID                     int64
	StartTime              string
	EndTime                string
	LightSleepSeconds      int
	DeepSleepSeconds       int
	RemSleepSeconds        int
	AwakeSeconds           int
	WakeUpCount            int
	DurationToSleepSeconds int
	DurationToWakeSeconds  int
	SnoringSeconds         int
	SnoringEpisodes        int
	AvgHeartRate           int
	MinHeartRate           int
	MaxHeartRate           int
	Source                 string
}

type BloodPressureRecord struct {
	ID        int64
	Timestamp string
	HeartRate int
	Systolic  int
	Diastolic int
	Comments  string
	Source    string
}

type HeartRateRecord struct {
	ID              int64
	Timestamp       string
	DurationSeconds string
	BpmValue        string
	Source          string
}
