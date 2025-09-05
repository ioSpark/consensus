package app

import (
	"fmt"
	"slices"
)

var ErrInvalidPoint = fmt.Errorf("invalid point value given")

type Point int

// TODO: Make into config file
// TODO: 0 will be the default if not voted? may cause problems
var PointValues = []Point{1, 2, 3, 5, 8, 13}

func NewPoint(value int) (Point, error) {
	if !slices.Contains(PointValues, Point(value)) {
		return -1, ErrInvalidPoint
	}

	return Point(value), nil
}
