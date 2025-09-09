package memory

import (
	"errors"
	"slices"

	"consensus/app"
)

func (r *Repository) Users() []app.UserID {
	// Probably a more efficient way of doing this
	s := make([]app.UserID, len(r.users))
	copy(s, r.users)
	return s
}

func (r *Repository) User(name string) (app.UserID, error) {
	for _, u := range r.users {
		if string(u) == name {
			return u, nil
		}
	}
	return "", app.ErrUserNotExist
}

func (r *Repository) CreateUser(u app.UserID) error {
	_, err := r.User(string(u))
	if err == nil {
		return app.ErrUserAlreadyExists
	} else if !errors.Is(err, app.ErrUserNotExist) {
		panic(err)
	}

	r.users = append(r.users, u)
	return nil
}

func (r *Repository) DeleteUser(ID app.UserID) error {
	_, err := r.User(string(ID))
	if errors.Is(err, app.ErrUserNotExist) {
		return err
	} else if err != nil {
		panic(err)
	}

	r.users = slices.DeleteFunc(r.users, func(u app.UserID) bool {
		return u == ID
	})

	return nil
}
