package http

import (
	"context"
	"net/http"
	"strings"

	"consensus/app"
	"consensus/html"

	"github.com/go-chi/chi/v5"

	g "maragu.dev/gomponents"
)

func indexHandler(
	w http.ResponseWriter,
	r *http.Request,
	s app.Storage,
) (g.Node, error) {
	user := r.Context().Value(contextUser).(app.User)
	return html.Index(html.PageProps{}, user, s.Tickets(), s.Users()), nil
}

// Return all revealed tickets as rows
func revealedHandler(
	w http.ResponseWriter,
	r *http.Request,
	s app.Storage,
) (g.Node, error) {
	user := r.Context().Value(contextUser).(app.User)

	// Return just the table if we're a HTMX request
	if r.Header.Get("HX-Request") == "true" {
		group := g.Group{}
		for _, t := range s.Tickets() {
			if !t.Revealed {
				continue
			}
			group = append(group, html.RevealedRow(*t, user))

		}
		return group, nil
	}
	return html.Revealed(html.PageProps{}, user, s.Tickets()), nil
}

func newTicketHandler(
	w http.ResponseWriter,
	r *http.Request,
	s app.Storage,
) (g.Node, error) {
	title := strings.TrimSpace(r.FormValue("title"))
	link := strings.TrimSpace(r.FormValue("link"))
	if title == "" || link == "" {
		// TODO: Return HTMX error
		return g.Text("title or link were blank"), httpError{http.StatusBadRequest}
	}

	user := r.Context().Value(contextUser).(app.User)
	_, err := s.CreateTicket(app.NewTicket(title, link, user))
	if err == app.ErrTicketAlreadyExists {
		// TODO: HTMX error
		// TODO: Might not need to exist, since all tickets are based on ID
		return g.Text(
				"ticket with name already exists",
			), httpError{
				http.StatusBadRequest,
			}
	} else if err != nil {
		// Only returns "already exists"
		panic(err)
	}

	return g.Group{
		html.ToPointPartial(user, s.Tickets(), s.Users()),
		html.InputRow(true),
	}, nil
}

func toPoint(w http.ResponseWriter, r *http.Request, s app.Storage) (g.Node, error) {
	user := r.Context().Value(contextUser).(app.User)

	if r.Header.Get("HX-Request") == "true" {
		group := g.Group{}
		for _, t := range s.Tickets() {
			if t.Revealed {
				continue
			}
			group = append(group, html.TicketRow(t, user, s.Users()))
		}
		return group, nil
	}
	return g.Text("you shouldn't be here"), httpError{http.StatusTeapot}
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
		r.Get("/to-point", provideStorage(s, toPoint))
	})
}
