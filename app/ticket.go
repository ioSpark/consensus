package app

import (
	"errors"
	"fmt"
	"slices"
)

var (
	ErrCantReveallNoVotes  = fmt.Errorf("nobody voted, cannot reveal")
	ErrTicketAlreadyExists = fmt.Errorf("ticket with name already exists")
	ErrTicketNotExist      = fmt.Errorf("ticket does not exist")
	ErrUserCantReveal      = fmt.Errorf("user did not raise ticket, cannot reveal")
)

// TODO: Determine what should be pointers
type TicketStorage interface {
	Ticket(ID int) (*Ticket, error)
	TicketByName(name string) (*Ticket, error)
	Tickets() []*Ticket
	// TODO: Should we accept parameters and create our own struct?
	CreateTicket(t Ticket) (*Ticket, error)
	DeleteTicket(ID int) error
	UpdateTicket(t Ticket) error
}

type Ticket struct {
	ID       int
	Name     string
	Link     string
	RaisedBy UserID
	Votes    map[UserID]Point
	Revealed bool
}

func (t *Ticket) Vote(userID UserID, v int) error {
	p, err := NewPoint(v)
	// Only returns ErrInvalidPoint
	if errors.Is(err, ErrInvalidPoint) {
		return err
	} else if err != nil {
		panic(err)
	}

	t.Votes[userID] = p
	return nil
}

func (t *Ticket) CanReveal(userID UserID) error {
	if userID != t.RaisedBy {
		return ErrUserCantReveal
	} else if len(t.Votes) == 0 {
		return ErrCantReveallNoVotes
	}
	return nil
}

func (t *Ticket) Reveal(userID UserID) error {
	err := t.CanReveal(userID)
	if err != nil {
		return err
	}
	t.Revealed = true
	return nil
}

// TODO: Should these be methods? or just helper functions in template funcmap?
//
//	Perhaps if we introduce non-numbers?
func (t *Ticket) Average() float64 {
	total := 0
	for _, v := range t.Votes {
		total += int(v)
	}
	return float64(total) / float64(len(t.Votes))
}

// Mode returns the most frequent point values, in ascending order.
// If there are no points, it returns nil.
func (t *Ticket) Mode() []Point {
	// Mode shouldn't be called if the ticket is not revealed (thus, must have points)
	// but handle it anyway.
	if len(t.Votes) == 0 {
		return nil
	}

	counts := make(map[Point]int, len(t.Votes))
	maxCount := 0

	for _, p := range t.Votes {
		c := counts[p] + 1
		counts[p] = c
		if c > maxCount {
			maxCount = c
		}
	}

	result := make([]Point, 0)
	for p, c := range counts {
		if c == maxCount {
			result = append(result, p)
		}
	}

	slices.Sort(result)
	return result
}

func (t *Ticket) Voted(userID UserID) bool {
	_, ok := t.Votes[userID]
	return ok
}

func NewTicket(name, link string, userID UserID) Ticket {
	return Ticket{
		Name:     name,
		Link:     link,
		RaisedBy: userID,
		Votes:    make(map[UserID]Point),
	}
}
