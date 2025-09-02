package app

import (
	"fmt"
	"slices"
)

var ErrInvalidPoint = fmt.Errorf("invalid point value given")

// TODO: Make into config file
// TODO: 0 will be the default if not voted? may cause problems
var PointValues = []int{1, 2, 3, 5, 8, 13}

type Point struct {
	User  User
	Point int
}

func NewPoint(user User, value int) (Point, error) {
	if !slices.Contains(PointValues, value) {
		return Point{}, ErrInvalidPoint
	}

	return Point{user, value}, nil
}

// Also updates existing point
func AddPoint(ticket *Ticket, user User, value int) (*Point, error) {
	p, err := NewPoint(user, value)
	// Only returns ErrInvalidPoint
	if err != nil {
		return &Point{}, err
	}

	ticket.Points[user] = p
	return &p, nil
}
