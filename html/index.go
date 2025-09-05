package html

import (
	"consensus/app"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	gh "maragu.dev/gomponents/html"
)

func Index(
	props PageProps,
	user app.User,
	tickets []*app.Ticket,
	users []*app.User,
) g.Node {
	props.Title = "Consensus"

	return page(
		props,
		user,
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
				hx.Trigger("every 5s"),
				hx.Get("/to-point"),
				g.Attr("_", hyperscriptTable),
				ToPointPartial(user, tickets, users),
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

		revealedTable(user, tickets),
	)
}

func Revealed(props PageProps, user app.User, tickets []*app.Ticket) g.Node {
	props.Title = "Consensus - Revealed"
	return page(
		props,
		user,
		revealedTable(user, tickets),
	)
}

func revealedTable(user app.User, tickets []*app.Ticket) g.Node {
	// TODO: Sort
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
			),
		),

		gh.TBody(
			hx.Trigger("newRevealed from:body, every 5s"),
			hx.Get("/revealed"),
			g.Attr("_", hyperscriptTable),
			g.Map(tickets, func(t *app.Ticket) g.Node {
				if t.Revealed {
					return RevealedRow(*t, user)
				}
				return g.Group{}
			}),
		),
	)
}

func ToPointPartial(user app.User, tickets []*app.Ticket, allUsers []*app.User) g.Node {
	// TODO: Sort
	return g.Group{
		g.Map(tickets, func(t *app.Ticket) g.Node {
			if !t.Revealed {
				return TicketRow(t, user, allUsers)
			}
			// TODO: Is there a better "zero" value?
			return g.Group{}
		}),
	}
}
