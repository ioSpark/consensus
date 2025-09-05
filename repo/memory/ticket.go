package memory

import (
	"errors"
	"math/rand/v2"
	"slices"

	"consensus/app"
)

func (r *Repository) Tickets() []*app.Ticket {
	return r.tickets
}

func (r *Repository) Ticket(ID int) (*app.Ticket, error) {
	for _, t := range r.Tickets() {
		if t.ID == ID {
			return t, nil
		}
	}
	return nil, app.ErrTicketNotExist
}

func (r *Repository) TicketByName(name string) (*app.Ticket, error) {
	for _, t := range r.Tickets() {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, app.ErrTicketNotExist
}

func (r *Repository) CreateTicket(t app.Ticket) (*app.Ticket, error) {
	// Generate un-used ID
	var newID int
	// Can be improved
	for {
		newID = rand.IntN(8192) // Enough space without IDs being unweidly
		_, err := r.Ticket(newID)
		if errors.Is(err, app.ErrTicketNotExist) {
			break
		} else if err != nil {
			panic(err)
		}
	}

	t.ID = newID

	r.tickets = append(r.tickets, &t)
	return &t, nil
}

func (r *Repository) DeleteTicket(ID int) error {
	_, err := r.Ticket(ID)
	if err != nil && err == app.ErrTicketNotExist {
		panic(err)
	}

	// Not the best way to do this
	r.tickets = slices.DeleteFunc(r.tickets, func(t *app.Ticket) bool {
		return t.ID == ID
	})

	return nil
}

func (r *Repository) UpdateTicket(t app.Ticket) error {
	_, err := r.Ticket(t.ID)
	if err != nil && err == app.ErrTicketNotExist {
		panic(err)
	}

	err = r.DeleteTicket(t.ID)
	if err != nil {
		panic(err)
	}

	newTicket, err := r.CreateTicket(t)
	if err != nil {
		panic(err)
	}

	newTicket.ID = t.ID

	return nil
}
