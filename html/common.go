package html

import (
	"fmt"

	"consensus/app"
	"consensus/build"

	g "maragu.dev/gomponents"
	gc "maragu.dev/gomponents/components"
	gh "maragu.dev/gomponents/html"
)

type PageProps struct {
	Title       string
	Description string
}

func page(props PageProps, userID app.UserID, children ...g.Node) g.Node {
	return gc.HTML5(gc.HTML5Props{
		Title:       props.Title,
		Description: props.Description,
		Language:    "en",
		Head: []g.Node{
			gh.Link(gh.Rel("icon"), gh.Href("/static/favicon.ico")),
			gh.Link(gh.Rel("stylesheet"), gh.Href("/static/style.css")),
			gh.Script(gh.Src("/static/htmx.min.js")),
			gh.Script(gh.Src("/static/_hyperscript.min.js")),
		},
		Body: []g.Node{
			gh.Class(
				"bg-linear-to-br m-auto max-w-max from-amber-200 to-teal-300 bg-fixed min-h-screen flex flex-col",
			),
			header(userID),
			gh.Main(
				gh.Class("flex flex-col flex-1"),
				g.Group(children),
			),
			footer(),
		},
	})
}

func header(u app.UserID) g.Node {
	return gh.Header(
		gh.Class("flex justify-between"),
		gh.H1(gh.Class("text-3xl font-bold"), g.Text("Consensus")),
		gh.Div(
			gh.Class("flex flex-col items-end"),
			gh.Span(
				gc.JoinAttrs(
					"class",
					gh.Class("font-bold"),
					userImage(u, true),
				),
			),
			gh.A(
				gh.Href("/oauth2/sign_out"),
				gh.Button(
					gh.Class(
						"cursor-pointer rounded bg-red-400 px-1 text-white hover:bg-red-500",
					),
					g.Text("Logout"),
				),
			),
		),
	)
}

func footer() g.Node {
	commitLink := fmt.Sprintf("%s/commit/%s", build.Repo, build.Commit)
	releaseLink := fmt.Sprintf("%s/releases/tag/%s", build.Repo, build.Version)

	return gh.Div(
		gh.Div(gh.Class("h-[25vh]")),

		// 0.3.1 (abcdef1)
		// 0.3.1 (abcdef1) DIRTY
		gh.Footer(
			gh.Class(
				"border-t-1 border-x-1 border-emerald-400 bg-emerald-300 sticky bottom-0 text-right",
			),
			gh.Span(
				gh.Class("text-blue-600 pr-1 underline"),
				gh.A(g.Text(build.Version), gh.Href(releaseLink)),
			),
			gh.Span(
				gh.Class("pr-1"),
				g.Text("("),
				gh.A(
					gh.Class("text-blue-600 underline"),
					g.Text(build.CommitShort),
					gh.Href(commitLink),
				),
				g.Text(")"),
			),
			g.If(build.Modified, gh.Span(gh.Class("pr-1"), g.Text("DIRTY"))),
		),
	)
}

func userImage(ID app.UserID, includeName bool) g.Node {
	return g.Group{
		gh.Class("inline-flex items-baseline"),
		gh.Img(
			gh.Class("size-5 self-center rounded-full"),
			gh.Src(fmt.Sprintf("https://github.com/%s.png?size=20", ID)),
		),
		g.If(includeName, gh.Span(gh.Class("mx-1"), g.Text(string(ID)))),
	}
}

// TODO: Determine if we need the tmp variable or not
const hyperscriptTable = `
	on htmx:beforeSwap if target is me and not detail.isError
		make a <tbody/> called tmp
		put event.detail.xhr.responseText into tmp.innerHTML

		set :oldIds to []
		for tr in <tr/> in me append tr.id to :oldIds end
		log 'old', :oldIds

		set :newIds to []
		for tr in <tr/> in tmp append tr.id to :newIds end
		log 'new', :newIds

		for tr in <tr/> in me
			if not (:newIds contain tr.id) add .leaving to tr end
		end
	end

	on htmx:afterSwap if target is me
		for tr in <tr/> in me
			if (:newIds contains tr.id) and not (:oldIds contains tr.id)
				add .entering to tr
			end
		end
	end

	on htmx:afterSettle if target is me
		for tr in .entering in me remove .entering from tr end
	end
`
