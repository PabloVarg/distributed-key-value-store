package raft

import (
	"bytes"
	"encoding/gob"
	"log/slog"
)

const (
	Put = iota
	Delete
)

type StoreAction struct {
	Action int
	Key    string
	Value  []byte
}

func EncodeAction(l *slog.Logger, a StoreAction) ([]byte, error) {
	b := new(bytes.Buffer)

	if err := gob.NewEncoder(b).Encode(a); err != nil {
		l.Error("error encoding raft action", "err", err)
		return nil, err
	}

	return b.Bytes(), nil
}
