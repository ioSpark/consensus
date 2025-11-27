package test

import (
	"testing"

	"consensus/app"
)

type repoTest struct {
	name string
	fn   func(t *testing.T, repo app.Repository)
}

var registry []repoTest

// registerRepoTest is called from init() in each test file
func registerRepoTest(name string, fn func(t *testing.T, repo app.Repository)) {
	registry = append(registry, repoTest{name, fn})
}

// Factory creates a fresh, isolated repository, and any teardown/cleanup logic.
type Factory interface {
	Setup(t *testing.T) app.Repository
	Teardown(t *testing.T)
}

// Run executes all registered repo tests against the given implementation
func Run(t *testing.T, factory Factory) {
	for _, tc := range registry {
		t.Run(tc.name, func(t *testing.T) {
			repo := factory.Setup(t)
			defer factory.Teardown(t)

			tc.fn(t, repo)
		})
	}
}
