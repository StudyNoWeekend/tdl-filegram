package telegram

import (
	"context"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"github.com/iyear/tdl/core/storage"
)

// boltStorage 基于 bbolt 实现 tdl 的 storage.Storage 接口，用于持久化 Telegram session
var _ storage.Storage = (*boltStorage)(nil)

type boltStorage struct {
	db     *bbolt.DB
	bucket []byte
}

func newBoltStorage(dataDir, namespace string) (*boltStorage, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	db, err := bbolt.Open(filepath.Join(dataDir, namespace), 0o600, nil)
	if err != nil {
		return nil, err
	}
	b := []byte(namespace)
	if err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(b)
		return err
	}); err != nil {
		return nil, err
	}
	return &boltStorage{db: db, bucket: b}, nil
}

func (s *boltStorage) Get(_ context.Context, key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(s.bucket).Get([]byte(key))
		if v == nil {
			return storage.ErrNotFound
		}
		val = append([]byte(nil), v...)
		return nil
	})
	return val, err
}

func (s *boltStorage) Set(_ context.Context, key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(s.bucket).Put([]byte(key), value)
	})
}

func (s *boltStorage) Delete(_ context.Context, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(s.bucket).Delete([]byte(key))
	})
}

func (s *boltStorage) Close() error { return s.db.Close() }
