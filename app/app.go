package app

import (
	"strconv"
	"strings"
)

type Repository interface {
	UserRepository
	TicketRepository
}

type Number interface {
	int | float64 | float32
}

// NumberFmt turns a Number into a string rounded to 3 decimal places, with any
// trailing zeros removed
func NumberFmt[T Number](num T) string {
	round := strconv.FormatFloat(float64(num), 'f', 3, 64)
	round = strings.TrimRight(round, "0")
	round = strings.TrimRight(round, ".")

	return round
}
