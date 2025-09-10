package memory

import (
	"sync"

	"consensus/app"
)

// Non-persistent in-memory repository
type Repository struct {
	tickets []app.Ticket
	users   []app.UserID
	sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{}
}
