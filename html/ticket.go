package html

import (
	"fmt"
	"maps"
	"slices"
	"strconv"

	"consensus/app"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	gc "maragu.dev/gomponents/components"
	gh "maragu.dev/gomponents/html"
)

func TicketRow(t *app.Ticket, u app.User, allUsers []*app.User) g.Node {
	canReveal := t.CanReveal(u)

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
					g.If(t.RaisedBy.Name == u.Name, gh.Class("font-bold")),
					userImage(t.RaisedBy.Name, true),
				),
			),
		),

		gh.Td(
			gh.Class("px-1"),
			gh.Div(
				gh.Class("flex justify-center gap-1"),
				g.Map(app.PointValues, func(v int) g.Node {
					userPoint := t.Points[u]
					voted := false
					if userPoint.Point == v {
						voted = true
					} else {
						voted = false
					}

					return voteButton(*t, v, voted)
				}),
			),
		),

		gh.Td(
			gh.Class("flex items-center gap-1 px-1"),
			g.Map(allUsers, func(u *app.User) g.Node {
				return gh.Span(gc.JoinAttrs(
					"class",
					g.If(!t.Voted(*u), gh.Class("opacity-30")),
					userImage(u.Name, false),
				))
			}),
		),

		gh.Td(
			gh.Class("px-1"),
			gh.Button(
				gc.JoinAttrs(
					"class",
					g.If(canReveal == nil, gh.Class("cursor-pointer")),
					g.If(canReveal == nil, gh.Class("cursor-not-allowed")),
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
		),
	)
}

func InputRow(oob bool) g.Node {
	inputClass := "px-1 border m-1 bg-neutral-50 focus:outline-none focus:border-emerald-500 focus:shadow-emerald-500 focus:shadow-lg/60"

	row := gh.Div(
		gh.Class("my-2 py-1 border-y w-full"),
		gh.ID("new-ticket"),
		hx.Target("#tickets"),
		g.If(oob, hx.SwapOOB("outerHTML")),

		gh.Form(
			gh.AutoComplete("off"),
			hx.Post("/new"),

			gh.H4(gh.Class("text-lg font-bold"), g.Text("New Ticket")),
			gh.Input(
				gh.Class(inputClass),
				gh.Type("text"),
				gh.Required(),
				gh.Placeholder("Issue Title"),
				gh.Name("title"),
			),
			gh.Input(
				gh.Class(inputClass),
				gh.Type("text"),
				gh.Required(),
				gh.Placeholder("https://link-to-issue.example"),
				gh.Name("link"),
			),
			gh.Button(
				gh.Class("bg-amber-100 hover:bg-amber-300 rounded px-1 cursor-pointer"),
				g.Text("Create"),
			),
		),
	)

	if oob {
		return gh.Template(row)
	}
	return row
}

func RevealedRow(t app.Ticket, u app.User) g.Node {
	// Show largest values first
	ticketPoints := slices.SortedStableFunc(
		maps.Values(t.Points),
		func(a, b app.Point) int {
			if a.Point < b.Point {
				return 1
			} else if a.Point > b.Point {
				return -1
			}
			return 0
		},
	)

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
					g.If(t.RaisedBy.Name == u.Name, gh.Class("font-bold")),
					userImage(t.RaisedBy.Name, true),
				),
			),
		),

		gh.Td(
			gh.Class("px-1"),
			g.Map(ticketPoints, func(p app.Point) g.Node {
				return gh.Div(
					gc.JoinAttrs(
						"class",
						gh.Class("flex justify-end gap-2"),
						g.If(p.User.Name == u.Name, gh.Class("font-bold")),
					),
					gh.Span(userImage(p.User.Name, true)),
					gh.Span(
						gh.Class("font-mono"),
						g.Textf("%d", p.Point),
					),
				)
			}),
		),

		gh.Td(
			gh.Class("px-1 font-mono"),
			g.Text(strconv.FormatFloat(t.Average(), 'f', -1, 64)),
		),

		gh.Td(
			gh.Class("px-1 font-mono"),
			gh.Div(
				gh.Class("flex place-content-center gap-1"),
				g.Map(t.Mode(), func(i int) g.Node {
					return gh.Span(g.Textf("%d", i))
				}),
			),
		),
	)
}

func voteButton(t app.Ticket, value int, voted bool) g.Node {
	return gh.Button(
		gc.JoinAttrs(
			"class",
			g.If(voted, gh.Class("bg-emerald-500")),
			g.If(!voted, gh.Class("cursor-pointer bg-emerald-50 hover:bg-slate-300")),
			gh.Class("rounded px-1 font-mono"),
		),

		hx.Put(fmt.Sprintf("/ticket/%d/point/%d", t.ID, value)),
		hx.Swap("outerHTML"),
		hx.Target("closest tr"),

		g.Textf("%d", value),
	)
}
