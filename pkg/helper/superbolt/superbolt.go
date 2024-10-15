package superbolt

import (
	"encoding/json"
	"errors"
	"io/fs"

	"go.etcd.io/bbolt"
)

type DB struct {
	db      *bbolt.DB
	options *Options
}

type Options struct {
	BBolt     *bbolt.Options
	Marshal   func(v interface{}) ([]byte, error)
	Unmarshal func(data []byte, v interface{}) error
}

func Open(path string, mode fs.FileMode, options *Options) (*DB, error) {
	var opts = &Options{}

	if options != nil {
		if options.BBolt != nil {
			opts.BBolt = options.BBolt
		} else {
			opts.BBolt = bbolt.DefaultOptions
		}

		if options.Marshal != nil {
			opts.Marshal = options.Marshal
		} else {
			opts.Marshal = json.Marshal
		}

		if options.Unmarshal != nil {
			opts.Unmarshal = options.Unmarshal
		} else {
			opts.Unmarshal = json.Unmarshal
		}
	} else {
		opts.BBolt = bbolt.DefaultOptions
		opts.Marshal = json.Marshal
		opts.Unmarshal = json.Unmarshal
	}

	db, err := bbolt.Open(path, mode, opts.BBolt)
	if err != nil {
		return nil, err
	}
	return &DB{db: db, options: opts}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Begin(writable bool) (*Tx, error) {
	tx, err := db.db.Begin(writable)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, options: db.options}, nil
}

var (
	ErrBucketNotFound     = bbolt.ErrBucketNotFound
	ErrBucketExists       = bbolt.ErrBucketExists
	ErrBucketNameRequired = bbolt.ErrBucketNameRequired
	ErrKeyRequired        = bbolt.ErrKeyRequired
	ErrKeyTooLarge        = bbolt.ErrKeyTooLarge
	ErrValueTooLarge      = bbolt.ErrValueTooLarge
	ErrIncompatibleValue  = bbolt.ErrIncompatibleValue
	ErrKeyExists          = errors.New("key exists")
)
