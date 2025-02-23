package api

import (
	"log/slog"
	"net/http"
	"time"

	"go.etcd.io/raft/v3"
)

// @title Key Value store API
// @version 1.0
// @description This API provides a simple interface for storing, retrieving, updating, and deleting key-value pairs. It supports basic CRUD operations, enabling clients to efficiently manage data. Keys are unique strings, and values can be any valid JSON object
func NewHTTPServer(l *slog.Logger, addr string, n raft.Node) *http.Server {
	mux := routes(l, n)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	return srv
}
