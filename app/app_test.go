package app

import (
	"testing"
)

func TestNumberFmt(t *testing.T) {
	numbers := map[float64]string{
		// Whole numbers are fine
		1:          "1",
		999:        "999",
		1234567890: "1234567890",
		123000:     "123000",
		// Negative whole
		-1234:  "-1234",
		-99999: "-99999",

		// 3 decimal places, no rounding
		1.123456789: "1.123",
		1.111111111: "1.111",
		1.9:         "1.9",
		1.999:       "1.999",
		999.999:     "999.999",
		// Rounding
		1.987654321: "1.988",
		1.99999:     "2",
		999.9999:    "1000",
		123.4007:    "123.401",
	}

	for k, v := range numbers {
		result := NumberFmt(k)
		if result != v {
			t.Errorf("expected %s, got %s", v, result)
		}
	}
}
