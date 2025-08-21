package http

import (
	"context"
	"net/http"
	"strings"

	"consensus/app"
	"consensus/html"

	"github.com/go-chi/chi/v5"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(contextUser).(app.User)

	err := html.Index(w, user, app.AllTickets())
	if err != nil {
		panic(err)
	}
}

// Return all revealed tickets as rows
func revealedHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(contextUser).(app.User)

	for _, t := range app.AllTickets() {
		if !t.Revealed {
			continue
		}

		err := html.RevealedRow(w, t, user)
		if err != nil {
			panic(err)
		}
	}
}

func newTicketHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.FormValue("title"))
	link := strings.TrimSpace(r.FormValue("link"))
	if title == "" || link == "" {
		// TODO: Return HTMX error
		http.Error(w, "title or link were blank", http.StatusBadRequest)
		return
	}

	user := r.Context().Value(contextUser).(app.User)
	t, err := app.NewTicket(title, link, user)
	if err == app.ErrTicketAlreadyExists {
		// TODO: HTMX error
		http.Error(w, "ticket with name already exists", http.StatusBadRequest)
		return
	} else if err != nil {
		// Only returns "already exists"
		panic(err)
	}

	err = html.TicketRow(w, t, user)
	if err != nil {
		panic(err)
	}
	err = html.NewTicketInput(w, true)
	if err != nil {
		panic(err)
	}
}

func userCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardedUser := r.Header.Get("X-Forwarded-User")
		if forwardedUser == "" {
			http.Error(
				w,
				"bypassed proxy?? no X-Forwarded-User header",
				http.StatusBadRequest,
			)
			return
		}

		u, err := app.GetUser(forwardedUser)
		if err == app.ErrUserNotExist {
			u = *app.NewUser(forwardedUser)
		}

		ctx := context.WithValue(r.Context(), contextUser, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Index(r chi.Router) {
	r.Route("/", func(r chi.Router) {
		r.Use(userCtx)

		r.Get("/", indexHandler)

		r.Post("/new", newTicketHandler)
		r.Get("/revealed", revealedHandler)
	})
}
