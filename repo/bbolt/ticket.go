package bbolt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"consensus/app"

	bolt "go.etcd.io/bbolt"
)

func ticketFromBucket(id int, b *bolt.Bucket) (app.Ticket, error) {
	var ticket app.Ticket

	ticket.Name = string(b.Get([]byte("name")))
	ticket.Link = string(b.Get([]byte("link")))

	ticket.RaisedBy = app.NewUser(string(b.Get([]byte("raisedBy"))))

	var err error

	ticket.CreatedAt, err = time.Parse(
		time.RFC3339Nano,
		string(b.Get([]byte("createdAt"))),
	)
	if err != nil {
		return app.Ticket{}, fmt.Errorf("parsing CreatedAt time: %v", err)
	}

	ticket.RevealedAt, err = time.Parse(
		time.RFC3339Nano,
		string(b.Get([]byte("revealedAt"))),
	)
	if err != nil {
		return app.Ticket{}, fmt.Errorf("parsing RevealedAt time: %v", err)
	}

	revealed := b.Get([]byte("revealed"))
	if revealed == nil {
		return app.Ticket{}, fmt.Errorf("revealed value not found in db")
	}
	switch string(revealed) {
	case "1":
		ticket.Revealed = true
	case "0":
		ticket.Revealed = false
	default:
		return app.Ticket{}, fmt.Errorf("expected 1 or 0 from db, got: %s", revealed)
	}

	votes := make(map[app.UserID]app.Point)
	voteBucket := b.Bucket([]byte("votes"))
	if voteBucket == nil {
		// TODO: Should we return an error anyway?
		panic(fmt.Sprintf("votes bucket does not exist for ticket %d, DB error?", id))
	}

	err = voteBucket.ForEach(func(k, v []byte) error {
		point, err := btoi(v)
		if err != nil {
			panic(fmt.Sprintf("couldn't parse stored point value: %v", v))
		}
		p, err := app.NewPoint(point)
		if err != nil {
			panic(fmt.Sprintf("stored point value not valid: %v", err))
		}

		votes[app.NewUser(string(k))] = p

		return nil
	})
	if err != nil {
		return app.Ticket{}, err
	}

	ticket.Votes = votes
	ticket.ID = id

	return ticket, nil
}

func ticketToBucket(ticket app.Ticket, b *bolt.Bucket) error {
	err := b.Put([]byte("name"), []byte(ticket.Name))
	if err != nil {
		return fmt.Errorf("storing name to db: %w", err)
	}
	err = b.Put([]byte("link"), []byte(ticket.Link))
	if err != nil {
		return fmt.Errorf("storing link to db: %w", err)
	}
	err = b.Put([]byte("raisedBy"), []byte(ticket.RaisedBy))
	if err != nil {
		return fmt.Errorf("storing raised by user to db: %w", err)
	}

	// TODO: Is there a better way for this?
	if ticket.Revealed {
		err = b.Put([]byte("revealed"), []byte("1"))
	} else {
		err = b.Put([]byte("revealed"), []byte("0"))
	}
	if err != nil {
		return fmt.Errorf("storing revealed value to DB: %w", err)
	}

	err = b.Put(
		[]byte("createdAt"),
		[]byte(ticket.CreatedAt.UTC().Format(time.RFC3339Nano)),
	)
	if err != nil {
		return fmt.Errorf("storing created at time to DB: %w", err)
	}
	err = b.Put(
		[]byte("revealedAt"),
		[]byte(ticket.RevealedAt.UTC().Format(time.RFC3339Nano)),
	)
	if err != nil {
		return fmt.Errorf("storing revealed at time to DB: %w", err)
	}

	voteBucket, err := b.CreateBucketIfNotExists([]byte("votes"))
	if err != nil {
		return fmt.Errorf("creating votes bucket: %v", err)
	}

	for k, v := range ticket.Votes {
		err := voteBucket.Put([]byte(k), itob(int(v)))
		if err != nil {
			return fmt.Errorf("storing vote %v into votes bucket: %w", v, err)
		}
	}

	return nil
}

func (r *Repository) Tickets() []app.Ticket {
	tickets := make([]app.Ticket, 0)

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket"))

		return b.ForEachBucket(func(k []byte) error {
			ticketBucket := b.Bucket(k)

			id, err := btoi(k)
			if err != nil {
				panic(err)
			}

			t, err := ticketFromBucket(id, ticketBucket)
			if err != nil {
				panic(err)
			}

			tickets = append(tickets, t)
			return nil
		})
	})
	if err != nil {
		panic(err)
	}

	return tickets
}

func (r *Repository) Ticket(id int) (app.Ticket, error) {
	var ticket app.Ticket
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket")).Bucket(itob(id))
		if b == nil {
			return app.ErrTicketNotExist
		}

		var err error
		ticket, err = ticketFromBucket(id, b)
		if err != nil {
			panic(err)
		}

		return nil
	})

	if errors.Is(err, app.ErrTicketNotExist) {
		return app.Ticket{}, err
	} else if err != nil {
		return app.Ticket{}, fmt.Errorf("ticket: %v", err)
	}

	return ticket, nil
}

func (r *Repository) Vote(id int, userID app.UserID, v int) (app.Ticket, error) {
	var ticket app.Ticket

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket")).Bucket(itob(id))
		if b == nil {
			return app.ErrTicketNotExist
		}

		var err error
		ticket, err = ticketFromBucket(id, b)
		if err != nil {
			panic(err)
		}

		p, err := app.NewPoint(v)
		if err != nil {
			return err
		}

		ticket.Votes[userID] = p

		err = ticketToBucket(ticket, b)
		if err != nil {
			return err
		}

		return nil
	})

	if errors.Is(err, app.ErrTicketNotExist) {
		return app.Ticket{}, err
	} else if err != nil {
		return app.Ticket{}, fmt.Errorf("vote: %v", err)
	}

	return ticket, nil
}

func (r *Repository) Reveal(id int, userID app.UserID) (app.Ticket, error) {
	var ticket app.Ticket

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket")).Bucket(itob(id))
		if b == nil {
			return app.ErrTicketNotExist
		}

		var err error

		ticket, err = ticketFromBucket(id, b)
		if err != nil {
			panic(err)
		}

		err = ticket.CanReveal(userID)
		if err != nil {
			return err
		}

		ticket.Revealed = true
		ticket.RevealedAt = time.Now().UTC()

		err = ticketToBucket(ticket, b)
		if err != nil {
			panic(err)
		}

		return nil
	})
	if err != nil {
		return app.Ticket{}, err
	}

	return ticket, nil
}

func (r *Repository) TicketByName(name string) (app.Ticket, error) {
	var ticket app.Ticket

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket"))

		// TODO: Break early when ticket found (use Cursor() instead of ForEach())
		return b.ForEachBucket(func(k []byte) error {
			ticketBucket := b.Bucket(k)

			id, err := btoi(k)
			if err != nil {
				panic(err)
			}

			t, err := ticketFromBucket(id, ticketBucket)
			if err != nil {
				panic(err)
			}

			if strings.EqualFold(t.Name, name) {
				ticket = t
				return nil
			}

			return nil
		})
	})

	if err != nil {
		return app.Ticket{}, fmt.Errorf("ticket by name: %v", err)
	} else if ticket.ID == 0 {
		return app.Ticket{}, app.ErrTicketNotExist
	}

	return ticket, nil
}

func (r *Repository) CreateTicket(
	name, link string,
	user app.UserID,
) (app.Ticket, error) {
	var ticket app.Ticket

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket"))

		userBucket := tx.Bucket([]byte("user")).Get([]byte(user))
		if userBucket == nil {
			return app.ErrUserNotExist
		}

		err := b.ForEachBucket(func(k []byte) error {
			ticketBucket := b.Bucket(k)

			// TODO: Check if nil?
			ticketName := string(ticketBucket.Get([]byte("name")))
			ticketLink := string(ticketBucket.Get([]byte("link")))

			if strings.EqualFold(ticketName, name) ||
				strings.EqualFold(ticketLink, link) {
				return app.ErrTicketAlreadyExists
			}

			return nil
		})
		if err != nil {
			return err
		}

		id, _ := b.NextSequence()
		t := app.NewTicket(int(id), name, link, user)

		ticketBucket, err := b.CreateBucket(itob(id))
		if err != nil {
			panic(err)
		}

		err = ticketToBucket(t, ticketBucket)
		if err != nil {
			panic(err)
		}

		ticket, err = ticketFromBucket(t.ID, ticketBucket)
		if err != nil {
			panic(err)
		}

		return nil
	})

	if errors.Is(err, app.ErrUserNotExist) ||
		errors.Is(err, app.ErrTicketAlreadyExists) {
		return app.Ticket{}, err
	} else if err != nil {
		return app.Ticket{}, fmt.Errorf("create ticket: %v", err)
	}

	return ticket, nil
}

func (r *Repository) DeleteTicket(id int) error {
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket"))

		ticketBucket := b.Bucket(itob(id))
		if ticketBucket == nil {
			return app.ErrTicketNotExist
		}

		return b.DeleteBucket(itob(id))
	})

	if errors.Is(err, app.ErrTicketNotExist) {
		return err
	} else if err != nil {
		return fmt.Errorf("delete ticket: %v", err)
	}
	return nil
}

func (r *Repository) UpdateTicket(id int, name, link string) (app.Ticket, error) {
	var ticket app.Ticket

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("ticket")).Bucket(itob(id))
		if b == nil {
			return app.ErrTicketNotExist
		}

		var err error
		ticket, err = ticketFromBucket(id, b)
		if err != nil {
			panic(err)
		}

		ticket.Name = name
		ticket.Link = link

		err = ticketToBucket(ticket, b)
		if err != nil {
			panic(err)
		}

		return nil
	})

	if errors.Is(err, app.ErrTicketNotExist) {
		return app.Ticket{}, err
	} else if err != nil {
		return app.Ticket{}, nil
	}

	return ticket, nil
}
