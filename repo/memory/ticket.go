package memory

import (
	"errors"
	"math/rand/v2"
	"slices"
	"strings"

	"consensus/app"
)

func (r *Repository) Tickets() []app.Ticket {
	// Probably a more efficient way of doing this
	s := make([]app.Ticket, len(r.tickets))
	copy(s, r.tickets)
	return s
}

func (r *Repository) Ticket(ID int) (app.Ticket, error) {
	for _, t := range r.tickets {
		if t.ID == ID {
			return t, nil
		}
	}
	return app.Ticket{}, app.ErrTicketNotExist
}

func (r *Repository) Vote(ID int, userID app.UserID, v int) (app.Ticket, error) {
	t, err := r.Ticket(ID)
	if errors.Is(err, app.ErrTicketNotExist) {
		return app.Ticket{}, err
	} else if err != nil {
		panic(err)
	}

	p, err := app.NewPoint(v)
	// Only returns ErrInvalidPoint
	if errors.Is(err, app.ErrInvalidPoint) {
		return app.Ticket{}, err
	} else if err != nil {
		panic(err)
	}

	t.Votes[userID] = p

	return t, nil
}

func (r *Repository) TicketByName(name string) (app.Ticket, error) {
	for _, t := range r.tickets {
		if t.Name == name {
			return t, nil
		}
	}
	return app.Ticket{}, app.ErrTicketNotExist
}

func (r *Repository) CreateTicket(ticket app.Ticket) (app.Ticket, error) {
	_, err := r.User(string(ticket.RaisedBy))
	if err != nil {
		if errors.Is(err, app.ErrUserNotExist) {
			return app.Ticket{}, err
		} else {
			panic(err)
		}
	}

	if slices.ContainsFunc(r.tickets, func(t app.Ticket) bool {
		if strings.EqualFold(t.Name, ticket.Name) {
			return true
		}
		return strings.EqualFold(t.Link, ticket.Link)
	}) {
		return app.Ticket{}, app.ErrTicketAlreadyExists
	}

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

	ticket.ID = newID

	r.tickets = append(r.tickets, ticket)
	return ticket, nil
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
	_, err := r.User(string(ticket.RaisedBy))
	if err != nil {
		if errors.Is(err, app.ErrUserNotExist) {
			return err
		} else {
			panic(err)
		}
	}

	_, err = r.Ticket(ticket.ID)
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
