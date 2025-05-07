package bboltstorage

import (
	"bytes"
	"encoding/gob"

	"github.com/bxcodec/httpcache/cache"
	"go.etcd.io/bbolt"
)

var bucketName = []byte("cache")

type Storage struct {
	db *bbolt.DB
}

func New(db *bbolt.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Set(key string, value cache.CachedResponse) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(value); err != nil {
		return err
	}

	if err := b.Put([]byte(key), buf.Bytes()); err != nil {
		return cache.ErrFailedToSaveToCache
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) get(key string) (*cache.CachedResponse, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(bucketName)
	if b == nil {
		return nil, cache.ErrCacheMissed
	}

	d := b.Get([]byte(key))
	if d == nil {
		return nil, cache.ErrCacheMissed
	}

	var res cache.CachedResponse
	if err := gob.NewDecoder(bytes.NewReader(d)).Decode(&res); err != nil {
		return nil, err
	}

	if err := tx.Rollback(); err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *Storage) Get(key string) (cache.CachedResponse, error) {
	res, err := s.get(key)
	if err != nil {
		return cache.CachedResponse{}, err
	}

	return *res, nil
}

func (s *Storage) Delete(key string) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return err
	}

	if err := b.Delete([]byte(key)); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Flush() error {
	return nil
}

func (s *Storage) Origin() string {
	return "bbolt"
}
