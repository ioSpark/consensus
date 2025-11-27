package bbolt_test

import (
	// "fmt"
	"os"
	"path/filepath"
	"testing"

	"consensus/app"
	"consensus/repo/bbolt"
	"consensus/repo/test"
)

type factory struct {
	dbPath string
}

func (f *factory) Setup(t *testing.T) app.Repository {
	tmp, err := os.MkdirTemp(os.TempDir(), "consensus-")
	if err != nil {
		t.Fatalf("could not create temp dir for bbolt: %v", err)
	}

	f.dbPath = filepath.Join(tmp, "test.db")

	r, err := bbolt.NewRepository(f.dbPath, bbolt.RepositoryOptions{})
	if err != nil {
		panic(err)
	}
	err = r.Initialise()
	if err != nil {
		panic(err)
	}

	return r
}

func (f *factory) Teardown(t *testing.T) {
	err := os.Remove(f.dbPath)
	if err != nil {
		t.Fatalf("could not remove temp db for bbolt: %v", err)
	}
	err = os.Remove(filepath.Dir(f.dbPath))
	if err != nil {
		t.Fatalf("could not remove temp dir for bbolt: %v", err)
	}
}

func TestBboltRepository(t *testing.T) {
	factory := factory{}

	test.Run(t, &factory)
}
