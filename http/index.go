package http

import (
	"context"
	"net/http"
	"strings"

	"consensus/app"
	"consensus/html"

	"github.com/go-chi/chi/v5"
)

func indexHandler(w http.ResponseWriter, r *http.Request, s app.Storage) {
	user := r.Context().Value(contextUser).(app.User)

	err := html.Index(w, user, s.Tickets(), s.Users())
	if err != nil {
		panic(err)
	}
}

// Return all revealed tickets as rows
func revealedHandler(w http.ResponseWriter, r *http.Request, s app.Storage) {
	user := r.Context().Value(contextUser).(app.User)

	// Return just the table if we're a HTMX request
	if r.Header.Get("HX-Request") == "true" {
		for _, t := range s.Tickets() {
			if !t.Revealed {
				continue
			}

			err := html.RevealedRow(w, t, user)
			if err != nil {
				panic(err)
			}
		}
	} else {
		err := html.Revealed(w, user, s.Tickets())
		if err != nil {
			panic(err)
		}
	}
}

func newTicketHandler(w http.ResponseWriter, r *http.Request, s app.Storage) {
	title := strings.TrimSpace(r.FormValue("title"))
	link := strings.TrimSpace(r.FormValue("link"))
	if title == "" || link == "" {
		// TODO: Return HTMX error
		http.Error(w, "title or link were blank", http.StatusBadRequest)
		return
	}

	user := r.Context().Value(contextUser).(app.User)
	t, err := s.CreateTicket(app.NewTicket(title, link, user))
	if err == app.ErrTicketAlreadyExists {
		// TODO: HTMX error
		http.Error(w, "ticket with name already exists", http.StatusBadRequest)
		return
	} else if err != nil {
		// Only returns "already exists"
		panic(err)
	}

	err = html.TicketRow(w, t, user, s.Users())
	if err != nil {
		panic(err)
	}
	err = html.NewTicketInput(w, true)
	if err != nil {
		panic(err)
	}
}

func userCtx(s app.Storage, next http.Handler) http.Handler {
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

		u, err := s.User(forwardedUser)
		if err == app.ErrUserNotExist {
			u = app.NewUser(forwardedUser)
			err := s.CreateUser(*u)
			if err != nil {
				panic(err)
			}
		}

		ctx := context.WithValue(r.Context(), contextUser, *u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Index(r chi.Router, s app.Storage) {
	r.Route("/", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return userCtx(s, next)
		})

		r.Get("/", provideStorage(s, indexHandler))

		r.Post("/new", provideStorage(s, newTicketHandler))
		r.Get("/revealed", provideStorage(s, revealedHandler))
	})
}
