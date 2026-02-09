package models

type ChartSeries struct {
	Label     string  `json:"label" gorm:"column:label"`
	Target    float64 `json:"target" gorm:"column:target"`
	Actual    float64 `json:"actual" gorm:"column:actual"`
	ActualOK  float64 `json:"actual_ok,omitempty" gorm:"column:actual_ok"`
	ActualNG  float64 `json:"actual_ng,omitempty" gorm:"column:actual_ng"`
	ExtraInfo string  `json:"extra_info,omitempty" gorm:"column:extra_info"`
}