package memory

import (
	"errors"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"

	"consensus/app"
)

// cloneTicket returns a safe copy of the provided ticket.
func cloneTicket(t app.Ticket) app.Ticket {
	clone := t
	if t.Votes != nil {
		clone.Votes = make(map[app.UserID]app.Point, len(t.Votes))
		maps.Copy(clone.Votes, t.Votes)
	}
	return clone
}

func (r *Repository) Tickets() []app.Ticket {
	r.RLock()
	defer r.RUnlock()

	dst := make([]app.Ticket, len(r.tickets))
	for i, t := range r.tickets {
		dst[i] = cloneTicket(t)
	}

	return dst
}

func (r *Repository) ticketWithoutLock(ID int) (app.Ticket, error) {
	for _, t := range r.tickets {
		if t.ID == ID {
			return cloneTicket(t), nil
		}
	}
	return app.Ticket{}, app.ErrTicketNotExist
}

func (r *Repository) Ticket(ID int) (app.Ticket, error) {
	r.RLock()
	defer r.RUnlock()

	return r.ticketWithoutLock(ID)
}

func (r *Repository) Vote(ID int, userID app.UserID, v int) (app.Ticket, error) {
	r.Lock()
	defer r.Unlock()

	t, err := r.ticketWithoutLock(ID)
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

	refreshed, err := r.updateTicketWithoutLock(t)
	if err != nil {
		return app.Ticket{}, err
	}

	return refreshed, nil
}

func (r *Repository) Reveal(ID int, userID app.UserID) (app.Ticket, error) {
	r.Lock()
	defer r.Unlock()

	ticket, err := r.ticketWithoutLock(ID)
	if err != nil {
		return app.Ticket{}, err
	}

	err = ticket.CanReveal(userID)
	if err != nil {
		return app.Ticket{}, err
	}
	ticket.Revealed = true

	updated, err := r.updateTicketWithoutLock(ticket)
	if err != nil {
		return app.Ticket{}, err
	}

	return updated, nil
}

func (r *Repository) TicketByName(name string) (app.Ticket, error) {
	r.RLock()
	defer r.RUnlock()
	for _, t := range r.tickets {
		if t.Name == name {
			return cloneTicket(t), nil
		}
	}
	return app.Ticket{}, app.ErrTicketNotExist
}

func (r *Repository) CreateTicket(ticket app.Ticket) (app.Ticket, error) {
	r.Lock()
	defer r.Unlock()

	_, err := r.userWithoutLock(string(ticket.RaisedBy))
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
		_, err := r.ticketWithoutLock(newID)
		if errors.Is(err, app.ErrTicketNotExist) {
			break
		} else if err != nil {
			panic(err)
		}
	}

	ticket.ID = newID

	r.tickets = append(r.tickets, ticket)
	return cloneTicket(ticket), nil
}

func (r *Repository) DeleteTicket(ID int) error {
	r.Lock()
	defer r.Unlock()

	_, err := r.ticketWithoutLock(ID)
	if errors.Is(err, app.ErrTicketNotExist) {
		return err
	} else if err != nil {
		panic(err)
	}

	r.tickets = slices.DeleteFunc(r.tickets, func(t app.Ticket) bool {
		return t.ID == ID
	})

	return nil
}

func (r *Repository) updateTicketWithoutLock(ticket app.Ticket) (app.Ticket, error) {
	_, err := r.userWithoutLock(string(ticket.RaisedBy))
	if err != nil {
		if errors.Is(err, app.ErrUserNotExist) {
			return app.Ticket{}, err
		} else {
			panic(err)
		}
	}

	_, err = r.ticketWithoutLock(ticket.ID)
	if errors.Is(err, app.ErrTicketNotExist) {
		return app.Ticket{}, err
	} else if err != nil {
		panic(err)
	}

	// The ticket we were provided could be out of date.
	// Try to merge where possible, but the provided ticket will take precedence.
	for i := range r.tickets {
		if r.tickets[i].ID == ticket.ID {
			mergedVotes := make(map[app.UserID]app.Point)
			maps.Copy(mergedVotes, r.tickets[i].Votes)
			maps.Copy(mergedVotes, ticket.Votes)

			r.tickets[i] = ticket
			r.tickets[i].Votes = mergedVotes

			return cloneTicket(r.tickets[i]), nil
		}
	}

	return app.Ticket{}, app.ErrTicketNotExist
}

func (r *Repository) UpdateTicket(ticket app.Ticket) (app.Ticket, error) {
	r.Lock()
	defer r.Unlock()

	return r.updateTicketWithoutLock(ticket)
}
