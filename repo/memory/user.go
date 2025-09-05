package memory

import (
	"slices"

	"consensus/app"
)

func (r *Repository) Users() []app.UserID {
	return r.users
}

func (r *Repository) User(name string) (app.UserID, error) {
	for _, u := range r.Users() {
		if string(u) == name {
			return u, nil
		}
	}
	return "", app.ErrUserNotExist
}

func (r *Repository) CreateUser(u app.UserID) error {
	_, err := r.User(string(u))
	if err != app.ErrUserNotExist && err != nil {
		panic(err)
	} else if err == nil {
		return app.ErrUserAlreadyExists
	}

	r.users = append(r.users, u)
	return nil
}

func (r *Repository) DeleteUser(ID app.UserID) error {
	_, err := r.User(string(ID))
	if err != nil && err == app.ErrUserNotExist {
		panic(err)
	}

	r.users = slices.DeleteFunc(r.users, func(u app.UserID) bool {
		return u == ID
	})

	return nil
}
