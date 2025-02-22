package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func readJSON(l *slog.Logger, r *http.Request, data any) error {
	defer func() {
		if err := r.Body.Close(); err != nil {
			l.Error("error closing request body", "path", r.URL.RawPath, "method", r.Method)
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		return err
	}

	return nil
}

func writeJSON(l *slog.Logger, data any, w http.ResponseWriter, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		l.Error("error writting json", "data", data)
		return err
	}

	return nil
}
