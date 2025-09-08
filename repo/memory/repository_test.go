package memory_test

import (
	"testing"

	"consensus/app"
	"consensus/repo/memory"
	"consensus/repo/test"
)

func TestMemoryRepository(t *testing.T) {
	test.Run(t, func(t *testing.T) app.Repository {
		return memory.NewRepository()
	})
}
