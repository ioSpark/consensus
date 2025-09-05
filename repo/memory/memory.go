package memory

import (
	"consensus/app"
)

// TODO: Unsafe, temporary in-memory "storage" layer. Could be coded better
type Repository struct {
	tickets []*app.Ticket
	users   []app.UserID
}

func NewRepository() *Repository {
	return &Repository{}
}
