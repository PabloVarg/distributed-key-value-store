package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"go.etcd.io/raft/v3"
)

func internalError(l *slog.Logger, r *http.Request, w http.ResponseWriter, err error) {
	l.Error("http internal error", "path", r.URL.String(), "method", r.Method, "error", err.Error())
	w.WriteHeader(http.StatusInternalServerError)
}

func badRequest(l *slog.Logger, w http.ResponseWriter, err error) {
	res := map[string]any{
		"error": err.Error(),
	}

	writeJSON(l, res, w, http.StatusBadRequest)
}

func unprocessableEntity(
	l *slog.Logger,
	w http.ResponseWriter,
	v validationResponse,
) {
	writeJSON(l, v, w, http.StatusUnprocessableEntity)
}

func RedirectToLeader(l *slog.Logger, w http.ResponseWriter, n raft.Node) {
	w.Header().Set("Location", fmt.Sprintf("%d", n.Status().Lead))
	w.WriteHeader(http.StatusTemporaryRedirect)
}
