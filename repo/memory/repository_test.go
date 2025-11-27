package memory_test

import (
	"testing"

	"consensus/app"
	"consensus/repo/memory"
	"consensus/repo/test"
)

type factory struct{}

func (f factory) Setup(_ *testing.T) app.Repository {
	return memory.NewRepository()
}

func (f factory) Teardown(_ *testing.T) {}

func TestMemoryRepository(t *testing.T) {
	test.Run(t, factory{})
}
