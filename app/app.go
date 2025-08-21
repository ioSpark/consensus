package app

import (
	"fmt"
	"maps"
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

type Ticket struct {
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

var ErrUserCantReveal = fmt.Errorf("user did not raise ticket, cannot reveal")

var ErrCantRevalNoVotes = fmt.Errorf("nobody voted, cannot reveal")

func (t *Ticket) CanReveal(user User) error {
	if user.Name != t.RaisedBy.Name {
		return ErrUserCantReveal
	} else if len(t.Points) == 0 {
		return ErrCantRevalNoVotes
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

// Most common value
// TODO: Cleanup
func (t *Ticket) Mode() []int {
	counts := make(map[int]int, len(PointValues))
	for _, p := range t.Points {
		counts[p.Point]++
	}

	reversed := make(map[int][]int, len(counts))
	for k, v := range counts {
		if reversed[v] == nil {
			reversed[v] = make([]int, 0)
		}
		reversed[v] = append(reversed[v], k)
	}

	ordered := slices.Collect(maps.Keys(reversed))
	slices.Sort(ordered)

	slices.Sort(reversed[ordered[0]])

	return reversed[ordered[0]]
}

var ErrTicketAlreadyExists = fmt.Errorf("ticket with name already exists")

func NewTicket(name, link string, user User) (*Ticket, error) {
	// Only returns not found err
	_, err := GetTicket(name)
	// Somthing unexpected went wrong
	// TODO: Better logic, this doesn't read well
	// If err is given, and it's not "not exist" then panic
	if err != ErrTicketNotExist && err != nil {
		panic(err)
	}
	if err == ErrTicketNotExist {
		m := make(map[User]Point, 0)
		t := Ticket{name, link, user, m, false}
		Tickets = append(Tickets, &t)
		return &t, nil
	}

	return &Ticket{}, ErrTicketAlreadyExists
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

type User struct {
	Name string
}

func NewUser(name string) *User {
	u := User{name}
	Users = append(Users, u)
	return &u
}

// TODO: Put these in DB for persistence
var (
	Tickets []*Ticket
	Users   []User
)

func AllTickets() []*Ticket {
	return Tickets
}

func AllUsers() []User {
	return Users
}

var ErrUserNotExist = fmt.Errorf("user does not exist")

func GetUser(name string) (User, error) {
	for _, u := range AllUsers() {
		if u.Name == name {
			return u, nil
		}
	}
	return User{}, ErrUserNotExist
}

var ErrTicketNotExist = fmt.Errorf("ticket does not exist")

func GetTicket(name string) (*Ticket, error) {
	for _, t := range AllTickets() {
		if t.Name == name {
			return t, nil
		}
	}
	return &Ticket{}, ErrTicketNotExist
}
