package api

import (
	"log/slog"
	"net/http"

	"go.etcd.io/raft/v3"

	_ "github.com/pablovarg/distributed-key-value-store/docs"
	"github.com/swaggo/http-swagger"
)

func routes(l *slog.Logger, n raft.Node) *http.ServeMux {
	mux := http.NewServeMux()

	all := hitLoggingMiddleware(l)
	mux.Handle("POST /values", all(NewPutHandler(l, n)))

	mux.HandleFunc(
		"/swagger-ui/",
		httpSwagger.Handler(),
	)

	return mux
}
