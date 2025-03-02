package store

import "errors"

var KeyNotFoundError = errors.New("key not found in store")

type Store struct {
	values map[string][]byte
}

func NewKeyValueStore() Store {
	return Store{
		values: make(map[string][]byte),
	}
}

func (s *Store) Put(key string, value []byte) {
	s.values[key] = value
}

func (s Store) Get(key string) ([]byte, error) {
	res, ok := s.values[key]
	if !ok {
		return nil, KeyNotFoundError
	}

	return res, nil
}
