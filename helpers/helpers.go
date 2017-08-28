package helpers

import (
	"time"
)

func AgePass(born, to time.Time) int {
	years := to.Year() - born.Year()
	if born.Month() > to.Month() {
		years--
	} else if born.Month() == to.Month() && born.Day() > to.Day() {
		years--
	}

	return years
}
