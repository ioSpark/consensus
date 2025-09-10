package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"consensus/app"
	"consensus/html"

	"github.com/go-chi/chi/v5"

	g "maragu.dev/gomponents"
)

func pointTicketHandler(
	w http.ResponseWriter,
	r *http.Request,
	repo app.Repository,
) (g.Node, error) {
	value, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
	if err != nil {
		// TODO: Return HTMX error
		return g.Textf(
				"point value is not an integer: %s",
				err,
			), httpError{
				http.StatusBadRequest,
			}
	}

	userID := r.Context().Value(contextUser).(app.UserID)
	ticket := r.Context().Value(contextTicket).(app.Ticket)

	// TODO: Is converting from int64 to int like this safe? (probably)
	votedTicket, err := repo.Vote(ticket.ID, userID, int(value))
	if err == app.ErrInvalidPoint {
		return g.Textf(
				"invalid point value %d",
				value,
			), httpError{
				http.StatusBadRequest,
			}
	} else if err != nil {
		panic(err)
	}

	return html.TicketRow(votedTicket, userID, repo.Users()), nil
}

func revealPointsHandler(
	w http.ResponseWriter,
	r *http.Request,
	repo app.Repository,
) (g.Node, error) {
	userID := r.Context().Value(contextUser).(app.UserID)
	ticket := r.Context().Value(contextTicket).(app.Ticket)

	// TODO: Return HTMX error
	_, err := repo.Reveal(ticket.ID, userID)
	if err == app.ErrUserCantReveal {
		return g.Text(
				"user did not raise ticket, cannot reveal",
			), httpError{
				http.StatusUnauthorized,
			}
	} else if err == app.ErrCantRevealNoVotes {
		return g.Text("no vots on ticket, cannot reveal"), httpError{http.StatusBadRequest}
	} else if err != nil {
		panic(err)
	}

	// The row will be deleted by providing no content.
	// This event triggers the client to reload the revealed table contents
	w.Header().Add("HX-Trigger", "newRevealed")
	return g.Group{}, nil
}

func ticketCtx(repo app.Repository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		IDstring := strings.TrimSpace(chi.URLParam(r, "ID"))

		// TODO: Determine if this is possible
		if IDstring == "" {
			// TODO: Probably 404?
			http.Error(w, "ticket name was blank", http.StatusBadRequest)
			return
		}

		ID, err := strconv.Atoi(IDstring)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("ticket ID was not integer: %s", err),
				http.StatusBadRequest,
			)
			return
		}

		t, err := repo.Ticket(ID)
		if err == app.ErrTicketNotExist {
			http.Error(
				w,
				fmt.Sprintf("ticket %d not found", ID),
				http.StatusBadRequest,
			)
			return
		} else if err != nil {
			// TODO: Return error
			panic(err)
		}

		ctx := context.WithValue(r.Context(), contextTicket, t)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Ticket(r chi.Router, s app.Repository) {
	r.Route("/ticket", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return userCtx(s, next)
		})
		r.Post("/", provideRepo(s, newTicketHandler))

		r.Route("/{ID}", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return ticketCtx(s, next)
			})

			r.Put("/point/{value}", provideRepo(s, pointTicketHandler))
			r.Post("/reveal", provideRepo(s, revealPointsHandler))
			// r.Delete("/", deleteTicketHandler)
		})
	})
}
