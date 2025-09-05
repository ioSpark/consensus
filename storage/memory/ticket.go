package memory

import (
	"errors"
	"math/rand/v2"
	"slices"

	"consensus/app"
)

func (s *Storage) Tickets() []*app.Ticket {
	return s.tickets
}

func (s *Storage) Ticket(ID int) (*app.Ticket, error) {
	for _, t := range s.Tickets() {
		if t.ID == ID {
			return t, nil
		}
	}
	return nil, app.ErrTicketNotExist
}

func (s *Storage) TicketByName(name string) (*app.Ticket, error) {
	for _, t := range s.Tickets() {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, app.ErrTicketNotExist
}

func (s *Storage) CreateTicket(t app.Ticket) (*app.Ticket, error) {
	// Generate un-used ID
	var newID int
	// Can be improved
	for {
		newID = rand.IntN(8192) // Enough space without IDs being unweidly
		_, err := s.Ticket(newID)
		if errors.Is(err, app.ErrTicketNotExist) {
			break
		} else if err != nil {
			panic(err)
		}
	}

	t.ID = newID

	s.tickets = append(s.tickets, &t)
	return &t, nil
}

func (s *Storage) DeleteTicket(ID int) error {
	_, err := s.Ticket(ID)
	if err != nil && err == app.ErrTicketNotExist {
		panic(err)
	}

	// Not the best way to do this
	s.tickets = slices.DeleteFunc(s.tickets, func(t *app.Ticket) bool {
		return t.ID == ID
	})

	return nil
}

func (s *Storage) UpdateTicket(t app.Ticket) error {
	_, err := s.Ticket(t.ID)
	if err != nil && err == app.ErrTicketNotExist {
		panic(err)
	}

	err = s.DeleteTicket(t.ID)
	if err != nil {
		panic(err)
	}

	newTicket, err := s.CreateTicket(t)
	if err != nil {
		panic(err)
	}

	newTicket.ID = t.ID

	return nil
}
