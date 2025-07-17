package analytics

import (
	"fmt"
)

type TimePeriod = int

const (
	Hour TimePeriod = iota + 1
	Day
	Week
	Month
	Year
	Date
	DateHour
)

func NewIncorrectPeriod(period int) error {
	return fmt.Errorf("expected TimePeriod enum, got %v", period)
}
