package api

import (
	"log/slog"
	"net/http"

	"go.etcd.io/raft/v3"
)

func routes(l *slog.Logger, n raft.Node) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /values", NewPutHandler(l, n))

	return mux
}
