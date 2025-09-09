package app

import (
	"fmt"
)

var (
	ErrUserNotExist      = fmt.Errorf("user does not exist")
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
)

type UserID string

// Currently the ID is just the GitHub user name
func NewUser(raw string) UserID {
	return UserID(raw)
}

type UserRepository interface {
	User(ID string) (UserID, error)
	Users() []UserID
	// TODO: Should we accept parameters and create our own struct?
	CreateUser(ID UserID) error
	DeleteUser(ID UserID) error
}
