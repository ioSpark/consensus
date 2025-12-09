package html

import (
	"fmt"
	"slices"
	"strconv"

	"consensus/app"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	gc "maragu.dev/gomponents/components"
	gh "maragu.dev/gomponents/html"
)

func TicketRow(t app.Ticket, userID app.UserID, allUsers []app.UserID) g.Node {
	canReveal := t.CanReveal(userID)
	isOwner := t.RaisedBy == userID

	return gh.Tr(
		gh.Class(
			"hover:bg-linear-to-r bg-emerald-200 hover:from-emerald-400 hover:to-teal-500 leaving:opacity-0 entering:opacity-0 transition-opacity duration-1000 opacity-100",
		),
		gh.ID(fmt.Sprintf("ticket-%d", t.ID)),
		gh.Td(
			gh.Class("px-1"),
			gh.A(
				gh.Class("text-blue-600 underline"),
				gh.Target("_blank"),
				gh.Href(t.Link),
				g.Text(t.Name),
			),
		),

		gh.Td(
			gh.Class("px-1"),
			gh.Span(
				gc.JoinAttrs(
					"class",
					g.If(t.RaisedBy == userID, gh.Class("font-bold")),
					userImage(t.RaisedBy, true),
				),
			),
		),

		gh.Td(
			gh.Class("px-1"),
			gh.Div(gh.Class("group relative"),
				gh.Div(gc.JoinAttrs(
					"class",
					g.If(!t.Voted(userID), gh.Class("hidden")),
					gh.Class(
						"z-10 inset-0 absolute bg-radial-[at_50%_0%] from-emerald-400 to-emerald-500 opacity-100 rounded duration-300 transition-opacity ease-out shadow-md/40 group-hover:opacity-0 htmx-swapping:opacity-0 group-hover:pointer-events-none",
					)),
				),

				gh.Div(
					gh.Class("flex flex-wrap justify-center gap-1"),
					g.Map(app.PointValues, func(v app.Point) g.Node {
						userVote := t.Votes[userID]
						voted := false
						if userVote == v {
							voted = true
						} else {
							voted = false
						}

						return voteButton(t.ID, v, voted)
					}),
				)),
		),

		gh.Td(
			// TODO: Ensure users are ordered by name
			gh.Div(
				gh.Class("flex flex-wrap justify-center gap-1 px-1"),
				g.Map(allUsers, func(u app.UserID) g.Node {
					return gh.Span(gc.JoinAttrs(
						"class",
						g.If(!t.Voted(u), gh.Class("opacity-30")),
						userImage(u, false),
					))
				}),
			),
		),

		gh.Td(
			gh.Div(
				gh.Class("px-1 flex flex-wrap gap-1 justify-center"),
				gh.Button(
					gc.JoinAttrs(
						"class",
						g.If(canReveal == nil, gh.Class("cursor-pointer")),
						g.If(canReveal != nil, gh.Class("cursor-not-allowed")),
						gh.Class(
							"rounded bg-amber-100 px-1 hover:bg-amber-300 disabled:bg-slate-100 disabled:opacity-50",
						),
					),

					g.If(canReveal == nil, g.Group{
						hx.Post(fmt.Sprintf("/ticket/%d/reveal", t.ID)),
						hx.Target("closest tr"),
						hx.Swap("outerHTML swap:1s"),
						// A successful response is to remove this row
						g.Attr(
							"_",
							"on htmx:afterOnLoad transition the closest <tr/> opacity to 0 over 1s",
						),
					}),
					g.If(canReveal != nil, gh.Disabled()),

					g.Text("Reveal"),
				),

				g.If(isOwner, rePointButton(t, userID)),
				g.If(isOwner, deleteButton(t, userID)),
			),
		),
	)
}

func InputRow(oob bool) g.Node {
	inputClass := "w-full px-1 border m-1 bg-neutral-50 focus:outline-none focus:border-emerald-500 focus:shadow-emerald-500 focus:shadow-lg/60"

	row := gh.Div(
		gh.Class("my-2 py-1 border-y w-full"),
		gh.ID("new-ticket"),
		hx.Target("#tickets"),
		g.If(oob, hx.SwapOOB("outerHTML")),

		gh.Form(
			gh.AutoComplete("off"),
			hx.Post("/ticket"),

			gh.H4(gh.Class("text-lg font-bold"), g.Text("New Ticket")),
			gh.Div(
				gh.Class("flex flex-row justify-between items-center"),
				gh.Div(
					gh.Class("w-full"),
					gh.Div(
						gh.Class("flex flex-row items-center"),
						gh.Label(g.Text("Title"), gh.For("ticket-title")),
						gh.Input(
							gh.ID("ticket-title"),
							gh.Class(inputClass),
							gh.Type("text"),
							gh.Required(),
							gh.Placeholder("Uh oh, Bug! 🐛"),
							gh.Name("title"),
						),
					),
					gh.Div(
						gh.Class("flex flex-row items-center"),
						gh.Label(g.Text("Link"), gh.For("ticket-link")),
						gh.Input(
							gh.ID("ticket-link"),
							gh.Class(inputClass),
							gh.Type("text"),
							gh.Required(),
							gh.Placeholder(
								"https://issue-tracker.example/bug-squash-123",
							),
							gh.Name("link"),
						),
					),
				),
				gh.Button(
					gh.Class(
						"min-h bg-amber-100 hover:bg-amber-300 rounded px-1 cursor-pointer",
					),
					g.Text("Create"),
				),
			),
		),
	)

	if oob {
		return gh.Template(row)
	}
	return row
}

func RevealedRow(t app.Ticket, userID app.UserID) g.Node {
	isOwner := t.RaisedBy == userID

	type vote struct {
		User  app.UserID
		Value app.Point
	}

	votes := make([]vote, 0)
	for u, v := range t.Votes {
		votes = append(votes, vote{u, v})
	}
	// Show largest values first, then by name.
	slices.SortStableFunc(votes, func(a, b vote) int {
		if a.Value < b.Value {
			return 1
		} else if a.Value > b.Value {
			return -1
		}

		if a.User < b.User {
			return 1
		} else if a.User > b.User {
			return -1
		}
		return 0
	})

	return gh.Tr(
		gh.Class(
			"hover:bg-linear-to-r bg-yellow-100 hover:from-yellow-200 hover:to-emerald-300 leaving:opacity-0 entering:opacity-0 transition-opacity duration-1000 opacity-100",
		),
		gh.ID(fmt.Sprintf("reveal-%d", t.ID)),
		gh.Td(
			gh.Class("px-1"),
			// TODO: Probably make links their own function
			gh.A(
				gh.Class("text-blue-600 underline"),
				gh.Target("_blank"),
				gh.Href(t.Link),
				g.Text(t.Name),
			),
		),

		gh.Td(
			gh.Class("px-1"),
			gh.Span(
				gc.JoinAttrs(
					"class",
					g.If(t.RaisedBy == userID, gh.Class("font-bold")),
					userImage(t.RaisedBy, true),
				),
			),
		),

		gh.Td(
			gh.Class("px-1"),
			g.Map(votes, func(v vote) g.Node {
				return gh.Div(
					gc.JoinAttrs(
						"class",
						gh.Class("flex justify-end gap-1"),
						g.If(v.User == userID, gh.Class("font-bold")),
					),
					gh.Span(userImage(v.User, true)),
					gh.Span(
						gh.Class("font-mono w-[2ch]"),
						g.Textf("%d", v.Value),
					),
				)
			}),
		),

		gh.Td(
			gh.Class("px-1 font-mono"),
			gh.Div(
				gh.Class("flex place-content-center"),
				g.Text(strconv.FormatFloat(t.Average(), 'f', -1, 64)),
			),
		),

		gh.Td(
			gh.Class("px-1 font-mono w-[2ch]"),
			gh.Div(
				gh.Class("flex flex-col place-content-center"),
				g.Map(t.Mode(), func(p app.Point) g.Node {
					return gh.Span(gh.Class("self-center"), g.Textf("%d", p))
				}),
			),
		),
		gh.Td(
			gh.Div(
				gh.Class("px-1 flex flex-wrap gap-1 justify-center"),
				g.If(isOwner, rePointButton(t, userID)),

				g.If(isOwner, deleteButton(t, userID)),
			),
		),
	)
}

func voteButton(tickedID int, point app.Point, voted bool) g.Node {
	return gh.Button(
		gc.JoinAttrs(
			"class",
			g.If(voted, gh.Class("bg-emerald-500")),
			g.If(!voted, gh.Class("cursor-pointer bg-emerald-50 hover:bg-slate-300")),
			gh.Class("rounded px-1 font-mono"),
		),

		hx.Put(fmt.Sprintf("/ticket/%d/point/%d", tickedID, point)),
		hx.Swap("outerHTML"),
		hx.Target("closest tr"),

		g.Textf("%d", point),
	)
}

func deleteButton(t app.Ticket, u app.UserID) g.Node {
	isOwner := t.RaisedBy == u

	return gh.Button(
		gc.JoinAttrs(
			"class",
			g.If(isOwner, gh.Class("cursor-pointer")),
			g.If(!isOwner, gh.Class("cursor-not-allowed")),
			gh.Class(
				"rounded bg-red-200 px-1 hover:bg-red-400 disabled:bg-slate-100 disabled:opacity-50",
			),
		),

		g.If(isOwner, g.Group{
			hx.Delete(fmt.Sprintf("/ticket/%d", t.ID)),
			hx.Target("closest tr"),
			hx.Swap("outerHTML swap:1s"),
			// A successful response is to remove this row
			g.Attr(
				"_",
				"on htmx:afterOnLoad transition the closest <tr/> opacity to 0 over 1s",
			),
		}),
		g.If(!isOwner, gh.Disabled()),

		g.Text("Delete"),
	)
}

func rePointButton(t app.Ticket, u app.UserID) g.Node {
	isPointed := len(t.Votes) > 0

	return gh.Button(
		gc.JoinAttrs(
			"class",
			g.If(isPointed, gh.Class("cursor-pointer")),
			g.If(!isPointed, gh.Class("cursor-not-allowed")),
			gh.Class(
				"rounded bg-emerald-100 px-1 hover:bg-emerald-400 disabled:bg-slate-100 disabled:opacity-50",
			),
		),
		g.If(!isPointed, gh.Disabled()),

		g.If(isPointed, g.Group{
			hx.Post(fmt.Sprintf("/ticket/%d/re-point", t.ID)),
			hx.Target("closest tr"),
			hx.Swap("outerHTML swap:1s"),
			// A successful response is to remove this row
			g.Attr(
				"_",
				"on htmx:afterOnLoad transition the closest <tr/> opacity to 0 over 1s",
			),
		}),
		g.Text("Re-Point"),
	)
}
