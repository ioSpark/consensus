package app

import (
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
	RaisedBy User
	// TODO: Maybe doesn't need to be a map? As Point contains user
	Points   map[User]Point
	Revealed bool
}

func (t *Ticket) Point(user User, v int) error {
	// TODO: Should we expect the caller to pass us a Point struct instead?
	p, err := NewPoint(user, v)
	// Only returns ErrInvalidPoint
	if err != nil {
		return err
	}

	t.Points[user] = p
	return nil
}

func (t *Ticket) CanReveal(user User) error {
	if user.Name != t.RaisedBy.Name {
		return ErrUserCantReveal
	} else if len(t.Points) == 0 {
		return ErrCantReveallNoVotes
	}
	return nil
}

func (t *Ticket) Reveal(user User) error {
	err := t.CanReveal(user)
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
	for _, v := range t.Points {
		total += v.Point
	}
	return float64(total) / float64(len(t.Points))
}

// Mode returns the most frequent point values, in ascending order.
// If there are no points, it returns nil.
func (t *Ticket) Mode() []int {
	// Mode shouldn't be called if the ticket is not revealed (thus, must have points)
	// but handle it anyway.
	if len(t.Points) == 0 {
		return nil
	}

	counts := make(map[int]int, len(t.Points))
	maxCount := 0

	for _, p := range t.Points {
		c := counts[p.Point] + 1
		counts[p.Point] = c
		if c > maxCount {
			maxCount = c
		}
	}

	result := make([]int, 0)
	for k, c := range counts {
		if c == maxCount {
			result = append(result, k)
		}
	}

	slices.Sort(result)
	return result
}

func (t *Ticket) Voted(user User) bool {
	for _, p := range t.Points {
		if p.User.Name == user.Name {
			return true
		}
	}
	return false
}

func NewTicket(name, link string, reporter User) Ticket {
	return Ticket{
		Name:     name,
		Link:     link,
		RaisedBy: reporter,
		Points:   make(map[User]Point),
	}
}
