package http

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"consensus/app"
	"consensus/html"

	"github.com/go-chi/chi/v5"

	g "maragu.dev/gomponents"
)

func indexHandler(
	w http.ResponseWriter,
	r *http.Request,
	repo app.Repository,
) (g.Node, error) {
	user := r.Context().Value(contextUser).(app.UserID)
	return html.Index(html.PageProps{}, user, repo.Tickets(), repo.Users()), nil
}

// Return all revealed tickets as rows
func revealedHandler(
	w http.ResponseWriter,
	r *http.Request,
	repo app.Repository,
) (g.Node, error) {
	userID := r.Context().Value(contextUser).(app.UserID)

	// TODO: Make revealed partial func
	tickets := repo.Tickets()
	slices.SortStableFunc(tickets, func(a, b app.Ticket) int {
		if a.RevealedAt.After(b.RevealedAt) {
			return 1
		}
		return -1
	})

	// Return just the table if we're a HTMX request
	if r.Header.Get("HX-Request") == "true" {
		group := g.Group{}
		for _, t := range tickets {
			if !t.Revealed {
				continue
			}
			group = append(group, html.RevealedRow(t, userID))

		}
		return group, nil
	}
	return html.Revealed(html.PageProps{}, userID, tickets), nil
}

func newTicketHandler(
	w http.ResponseWriter,
	r *http.Request,
	repo app.Repository,
) (g.Node, error) {
	title := strings.TrimSpace(r.FormValue("title"))
	link := strings.TrimSpace(r.FormValue("link"))
	if title == "" || link == "" {
		// TODO: Return HTMX error
		return g.Text("title or link were blank"), httpError{http.StatusBadRequest}
	}

	userID := r.Context().Value(contextUser).(app.UserID)
	_, err := repo.CreateTicket(title, link, userID)
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
		html.ToPointPartial(userID, repo.Tickets(), repo.Users()),
		html.InputRow(true),
	}, nil
}

func toPoint(
	w http.ResponseWriter,
	r *http.Request,
	repo app.Repository,
) (g.Node, error) {
	userID := r.Context().Value(contextUser).(app.UserID)

	if r.Header.Get("HX-Request") == "true" {
		group := g.Group{}
		for _, t := range repo.Tickets() {
			if t.Revealed {
				continue
			}
			group = append(group, html.TicketRow(t, userID, repo.Users()))
		}
		return group, nil
	}
	return g.Text("you shouldn't be here"), httpError{http.StatusTeapot}
}

func userCtx(s app.Repository, next http.Handler) http.Handler {
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
			err := s.CreateUser(u)
			if err != nil {
				panic(err)
			}
		}

		ctx := context.WithValue(r.Context(), contextUser, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Index(r chi.Router, repo app.Repository) {
	r.Route("/", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return userCtx(repo, next)
		})

		r.Get("/", provideRepo(repo, indexHandler))

		r.Get("/revealed", provideRepo(repo, revealedHandler))
		r.Get("/to-point", provideRepo(repo, toPoint))
	})
}
