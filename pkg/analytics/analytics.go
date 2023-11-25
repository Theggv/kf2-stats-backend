package analytics

import (
	"errors"
	"fmt"
)

type TimePeriod = int

const (
	Hour TimePeriod = iota + 1
	Day
	Week
	Month
	Year
)

func NewIncorrectPeriod(period int) error {
	return errors.New(fmt.Sprintf("expected TimePeriod enum, got %v", period))
}
