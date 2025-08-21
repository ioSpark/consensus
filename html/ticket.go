package html

import (
	"io"

	"consensus/app"
)

func TicketRow(w io.Writer, t *app.Ticket, u app.User) error {
	data := struct {
		Ticket *app.Ticket
		User   app.User
	}{
		Ticket: t,
		User:   u,
	}

	err := templates.ExecuteTemplate(w, "ticket-row.html", data)
	if err != nil {
		return err
	}
	return nil
}

func NewTicketInput(w io.Writer, oob bool) error {
	err := templates.ExecuteTemplate(w, "new-input.html", oob)
	if err != nil {
		return err
	}
	return nil
}

func RevealedRow(w io.Writer, t *app.Ticket, u app.User) error {
	data := struct {
		Ticket *app.Ticket
		User   app.User
	}{
		Ticket: t,
		User:   u,
	}

	err := templates.ExecuteTemplate(w, "revealed-row.html", data)
	if err != nil {
		return err
	}
	return nil
}
