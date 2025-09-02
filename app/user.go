package app

import (
	"fmt"
)

var (
	ErrUserNotExist      = fmt.Errorf("user does not exist")
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
)

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{name}
}

// TODO: Determine what should be pointers
type UserStorage interface {
	User(name string) (*User, error)
	Users() []*User
	CreateUser(user User) error
	DeleteUser(name string) error
}
