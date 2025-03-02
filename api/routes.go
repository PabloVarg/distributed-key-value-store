package api

import (
	"log/slog"
	"net/http"

	"go.etcd.io/raft/v3"

	_ "github.com/pablovarg/distributed-key-value-store/docs"
	"github.com/pablovarg/distributed-key-value-store/store"
	"github.com/swaggo/http-swagger"
)

func routes(l *slog.Logger, n raft.Node, s store.Store) *http.ServeMux {
	mux := http.NewServeMux()

	all := hitLoggingMiddleware(l)
	mux.Handle("POST /values", all(NewPutHandler(l, n)))
	mux.Handle("GET /values/{key}", all(NewGetHandler(l, n, s)))
	mux.Handle("DELETE /values/{key}", all(NewDeleteHandler(l, n)))

	mux.HandleFunc(
		"/swagger-ui/",
		httpSwagger.Handler(),
	)

	return mux
}
