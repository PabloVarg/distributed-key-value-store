package api

import (
	"log/slog"
	"net/http"

	"go.etcd.io/raft/v3"
)

func routes(l *slog.Logger, n raft.Node) *http.ServeMux {
	mux := http.NewServeMux()

	all := hitLoggingMiddleware(l)
	mux.Handle("POST /values", all(NewPutHandler(l, n)))

	return mux
}
