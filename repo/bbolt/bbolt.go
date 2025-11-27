package bbolt

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Repository struct {
	db   *bolt.DB
	log  *slog.Logger
	path string
}

type RepositoryOptions struct {
	Log *slog.Logger
}

type number interface {
	int | int64 | uint | uint64
}

func NewRepository(path string, opts RepositoryOptions) (*Repository, error) {
	if opts.Log == nil {
		opts.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	return &Repository{
		db:   db,
		log:  opts.Log,
		path: path,
	}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

// Initialise creates the user and ticket buckets if they don't exist already.
func (r *Repository) Initialise() error {
	err := r.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("user"))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("ticket"))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("initialising db: %w", err)
	}

	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob[T number](v T) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// btoi returns the int64 of the given bytes.
func btoi(b []byte) (int, error) {
	i := binary.BigEndian.Uint64(b)
	return int(i), nil
}

// // btoi returns the int64 of the given bytes.
// func btoi64(b []byte) (int64, error) {
// 	i := binary.BigEndian.Uint64(b)
// 	return int64(i), nil
// }
