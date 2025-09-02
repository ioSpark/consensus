package html

import (
	"embed"
	"errors"
	"html/template"
	"maps"
	"net/url"
	"slices"

	"consensus/app"
)

//go:embed template
var templateFS embed.FS

var templates *template.Template

func templateHelpers() template.FuncMap {
	return template.FuncMap{
		// https://stackoverflow.com/a/18276968 - with some adjustments
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"mkSlice":     func(values ...any) []any { return values },
		"pointValues": func() []int { return app.PointValues },
		"hasVoted": func(t *app.Ticket, user app.User) bool {
			for u := range t.Points {
				if u.Name == user.Name {
					return true
				}
			}
			return false
		},
		"sortPoints": func(p map[app.User]app.Point) []app.Point {
			points := slices.Collect(maps.Values(p))
			slices.SortStableFunc(points, func(a, b app.Point) int {
				if a.Point > b.Point {
					return 1
				} else if a.Point < b.Point {
					return -1
				}
				return 0
			})
			return points
		},
		"urlPathEscape": func(s string) string { return url.PathEscape(s) },
	}
}

// TODO: Is this fine?
func init() {
	var err error
	t := template.New("").Funcs(templateHelpers())
	templates, err = t.ParseFS(templateFS, "template/*")
	if err != nil {
		panic(err)
	}
}
