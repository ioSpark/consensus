package memory

import (
	"consensus/app"
)

// TODO: Unsafe, temporary in-memory "storage" layer. Could be coded better
type Storage struct {
	tickets []*app.Ticket
	users   []app.UserID
}

func NewStorage() *Storage {
	return &Storage{}
}
