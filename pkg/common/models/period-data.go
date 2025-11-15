package models

type PeriodData struct {
	Period string `json:"period"`

	Value         float64 `json:"value"`
	PreviousValue float64 `json:"prev"`
	Difference    float64 `json:"diff"`
	MaxValue      float64 `json:"max_value"`
	Trend         float64 `json:"trend"`
}
