package html

import (
	"io"

	"consensus/app"
)

func Index(w io.Writer, user app.User, tickets []*app.Ticket, users []*app.User) error {
	data := struct {
		User     app.User
		Tickets  []*app.Ticket
		AllUsers []*app.User
	}{
		User:     user,
		Tickets:  tickets,
		AllUsers: users,
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		return err
	}
	return nil
}

func Revealed(w io.Writer, user app.User, tickets []*app.Ticket) error {
	data := struct {
		User    app.User
		Tickets []*app.Ticket
	}{
		User:    user,
		Tickets: tickets,
	}

	err := templates.ExecuteTemplate(w, "revealed.html", data)
	if err != nil {
		return err
	}
	return nil
}
