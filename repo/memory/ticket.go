package memory

import (
	"errors"
	"math/rand/v2"
	"slices"

	"consensus/app"
)

func (r *Repository) Tickets() []app.Ticket {
	return r.tickets
}

func (r *Repository) Ticket(ID int) (app.Ticket, error) {
	for _, t := range r.Tickets() {
		if t.ID == ID {
			return t, nil
		}
	}
	return app.Ticket{}, app.ErrTicketNotExist
}

func (r *Repository) TicketByName(name string) (app.Ticket, error) {
	for _, t := range r.Tickets() {
		if t.Name == name {
			return t, nil
		}
	}
	return app.Ticket{}, app.ErrTicketNotExist
}

func (r *Repository) CreateTicket(t app.Ticket) (app.Ticket, error) {
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

	r.tickets = append(r.tickets, t)
	return t, nil
}

func (r *Repository) DeleteTicket(ID int) error {
	_, err := r.Ticket(ID)
	if errors.Is(err, app.ErrTicketNotExist) {
		return err
	} else if err != nil {
		panic(err)
	}

	// Not the best way to do this
	r.tickets = slices.DeleteFunc(r.tickets, func(t app.Ticket) bool {
		return t.ID == ID
	})

	return nil
}

func (r *Repository) UpdateTicket(ticket app.Ticket) error {
	_, err := r.Ticket(ticket.ID)
	if errors.Is(err, app.ErrTicketNotExist) {
		return err
	} else if err != nil {
		panic(err)
	}

	// Maybe a better way of doing this
	for i, t := range r.tickets {
		if t.ID == ticket.ID {
			r.tickets[i] = ticket
		}
	}

	return nil
}
