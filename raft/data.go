package raft

import (
	"bytes"
	"encoding/gob"
	"log/slog"
)

const (
	Put = iota
	Get
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

func DecodeAction(l *slog.Logger, data []byte) (StoreAction, error) {
	var res StoreAction

	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&res); err != nil {
		l.Error("error decoding raft action", "err", err)
		return StoreAction{}, err
	}

	return res, nil
}
