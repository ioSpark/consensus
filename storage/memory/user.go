package memory

import (
	"slices"

	"consensus/app"
)

func (s *Storage) Users() []app.UserID {
	return s.users
}

func (s *Storage) User(name string) (app.UserID, error) {
	for _, u := range s.Users() {
		if string(u) == name {
			return u, nil
		}
	}
	return "", app.ErrUserNotExist
}

func (s *Storage) CreateUser(u app.UserID) error {
	_, err := s.User(string(u))
	if err != app.ErrUserNotExist && err != nil {
		panic(err)
	} else if err == nil {
		return app.ErrUserAlreadyExists
	}

	s.users = append(s.users, u)
	return nil
}

func (s *Storage) DeleteUser(ID app.UserID) error {
	_, err := s.User(string(ID))
	if err != nil && err == app.ErrUserNotExist {
		panic(err)
	}

	s.users = slices.DeleteFunc(s.users, func(u app.UserID) bool {
		return u == ID
	})

	return nil
}
