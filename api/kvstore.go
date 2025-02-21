package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.etcd.io/raft/v3"
)

func NewPutHandler(l *slog.Logger, n raft.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		n.Propose(ctx, []byte("hello"))
	})
}
