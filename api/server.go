package api

import (
	"log/slog"
	"net/http"
	"time"

	"go.etcd.io/raft/v3"
)

func NewHTTPServer(l *slog.Logger, addr string, n raft.Node) *http.Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      routes(l, n),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	return srv
}
