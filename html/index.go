package html

import (
	"slices"

	"consensus/app"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	gh "maragu.dev/gomponents/html"
)

func Index(
	props PageProps,
	userID app.UserID,
	tickets []app.Ticket,
	users []app.UserID,
) g.Node {
	props.Title = "Consensus"

	return page(
		props,
		userID,
		gh.Table(
			gh.Class(
				"w-full border-separate border-spacing-y-1 border border-transparent",
			),
			gh.THead(
				gh.Tr(
					gh.Class("bg-emerald-400"),
					gh.Th(g.Text("Ticket")),
					gh.Th(g.Text("Reporter")),
					gh.Th(g.Text("Point Select")),
					gh.Th(g.Text("Voters")),
					gh.Th(),
				),
			),
			gh.TBody(
				gh.ID("tickets"),
				hx.Trigger("newToPoint from:body, every 5s"),
				hx.Get("/to-point"),
				g.Attr("_", hyperscriptTable),
				ToPointPartial(userID, tickets, users),
			),
		),

		InputRow(false),

		gh.A(
			gh.Href("/revealed"),
			gh.Button(
				gh.Class("cursor-pointer rounded bg-amber-100 px-1 hover:bg-amber-300"),
				g.Text("View Revealed Only"),
			),
		),

		revealedTable(userID, tickets),
	)
}

func Revealed(props PageProps, userID app.UserID, tickets []app.Ticket) g.Node {
	props.Title = "Consensus - Revealed"
	return page(
		props,
		userID,
		revealedTable(userID, tickets),
	)
}

func revealedTable(user app.UserID, tickets []app.Ticket) g.Node {
	slices.SortStableFunc(tickets, func(a, b app.Ticket) int {
		if a.RevealedAt.After(b.RevealedAt) {
			return 1
		}
		return -1
	})

	return gh.Table(
		gh.Class("w-full border-separate border-spacing-y-1 border border-transparent"),

		gh.THead(
			gh.Tr(
				gh.Class("bg-emerald-400"),
				gh.Th(g.Text("Ticket")),
				gh.Th(g.Text("Reporter")),
				gh.Th(g.Text("Voters")),
				gh.Th(g.Text("Avg")),
				gh.Th(g.Text("Mode")),
				gh.Th(),
			),
		),

		gh.TBody(
			hx.Trigger("newRevealed from:body, every 5s"),
			hx.Get("/revealed"),
			g.Attr("_", hyperscriptTable),
			g.Map(tickets, func(t app.Ticket) g.Node {
				if t.Revealed {
					return RevealedRow(t, user)
				}
				return g.Group{}
			}),
		),
	)
}

func ToPointPartial(
	user app.UserID,
	tickets []app.Ticket,
	allUsers []app.UserID,
) g.Node {
	slices.SortStableFunc(tickets, func(a, b app.Ticket) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return 1
		}
		return -1
	})

	return g.Group{
		g.Map(tickets, func(t app.Ticket) g.Node {
			if !t.Revealed {
				return TicketRow(t, user, allUsers)
			}
			// TODO: Is there a better "zero" value?
			return g.Group{}
		}),
	}
}
