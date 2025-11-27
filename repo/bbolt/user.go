package bbolt

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"consensus/app"

	bolt "go.etcd.io/bbolt"
)

// Also there's no real need to perform serialsation on the user type, in the future
// we will likely want to add other fields to the Uses. This prevents the need to
// perform a storage format migration.

func deserialiseUser(b []byte) (app.UserID, error) {
	var userID app.UserID

	d := gob.NewDecoder(bytes.NewReader(b))
	err := d.Decode(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func serialiseUser(userID app.UserID) ([]byte, error) {
	var buf bytes.Buffer

	e := gob.NewEncoder(&buf)
	err := e.Encode(userID)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func (r *Repository) Users() []app.UserID {
	users := make([]app.UserID, 0)

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))

		_ = b.ForEach(func(k, v []byte) error {
			u, err := deserialiseUser(v)
			if err != nil {
				panic(err)
			}

			users = append(users, u)
			return nil
		})

		return nil
	})
	if err != nil {
		panic(fmt.Errorf("users: %v", err))
	}

	return users
}

func (r *Repository) User(name string) (app.UserID, error) {
	var user app.UserID

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))

		userBytes := b.Get([]byte(name))
		if userBytes == nil {
			return app.ErrUserNotExist
		}

		var err error
		user, err = deserialiseUser(userBytes)
		if err != nil {
			panic(err)
		}

		return nil
	})

	if errors.Is(err, app.ErrUserNotExist) {
		return "", err
	} else if err != nil {
		return "", fmt.Errorf("get user: %v", err)
	}

	return user, nil
}

func (r *Repository) CreateUser(u app.UserID) error {
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))

		user := b.Get([]byte(u))
		if user != nil {
			return app.ErrUserAlreadyExists
		}
		userBytes, err := serialiseUser(u)
		if err != nil {
			panic(err)
		}
		// Users don't currently store any other data
		return b.Put([]byte(u), userBytes)
	})

	if errors.Is(err, app.ErrUserAlreadyExists) {
		return err
	} else if err != nil {
		return fmt.Errorf("create user: %v", err)
	}
	return nil
}

func (r *Repository) DeleteUser(ID app.UserID) error {
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))

		u := b.Get([]byte(ID))
		if u == nil {
			return app.ErrUserNotExist
		}

		return b.Delete([]byte(ID))
	})

	if errors.Is(err, app.ErrUserNotExist) {
		return err
	} else if err != nil {
		return fmt.Errorf("delete user: %v", err)
	}

	return nil
}
