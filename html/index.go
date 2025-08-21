package html

import (
	"io"

	"consensus/app"
)

func Index(w io.Writer, user app.User, tickets []*app.Ticket) error {
	data := struct {
		User    app.User
		Tickets []*app.Ticket
	}{
		User:    user,
		Tickets: tickets,
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		return err
	}
	return nil
}
