package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	internalRaft "github.com/pablovarg/distributed-key-value-store/raft"
	"github.com/pablovarg/distributed-key-value-store/store"
	"go.etcd.io/raft/v3"
)

// @title Put
// @description inserts or updates a key's value
// @accept json
// @param input body api.NewPutHandler.input true "Key / Value pair"
// @success 201
// @router /values [post]
func NewPutHandler(l *slog.Logger, n raft.Node) http.Handler {
	type input struct {
		Key   *string `json:"key"   validate:"required"`
		Value []byte  `json:"value" validate:"required" swaggertype:"string" format:"base64"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in input
		if err := readJSON(l, r, &in); err != nil {
			badRequest(l, w, err)
			return
		}

		if !internalRaft.IsLeader(n) {
			RedirectToLeader(l, w, n)
			return
		}

		v := validator.New(validator.WithRequiredStructEnabled())
		if err := v.Struct(in); err != nil {
			var vError validator.ValidationErrors

			switch {
			case errors.As(err, &vError):
				unprocessableEntity(l, w, buildErrorsResponse(vError))
			default:
				internalError(l, r, w, err)
			}
			return
		}

		a, err := internalRaft.EncodeAction(l, internalRaft.StoreAction{
			Action: internalRaft.Put,
			Key:    *in.Key,
			Value:  in.Value,
		})
		if err != nil {
			internalError(l, r, w, err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := n.Propose(ctx, a); err != nil {
			internalError(l, r, w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

// @title Get
// @description retrieves a key's value
// @param query path string true "key"
// @success 200
// @router /values/{key} [get]
func NewGetHandler(l *slog.Logger, n raft.Node, s store.Store) http.Handler {
	type output struct {
		Key   string `json:"key"`
		Value []byte `json:"value"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		if !internalRaft.IsLeader(n) {
			RedirectToLeader(l, w, n)
			return
		}

		value, err := s.Get(key)
		if err != nil {
			switch {
			case errors.Is(err, store.KeyNotFoundError):
				http.NotFound(w, r)
			default:
				internalError(l, r, w, err)
			}
			return
		}

		entry := output{
			Key:   key,
			Value: value,
		}

		writeJSON(l, entry, w, http.StatusOK)
	})
}

// @title Delete
// @description deletes a key, value pair from the store
// @param query path string true "key"
// @success 200
// @router /values/{key} [delete]
func NewDeleteHandler(l *slog.Logger, n raft.Node) http.Handler {
	type output struct {
		Key   string `json:"key"`
		Value []byte `json:"value"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		if !internalRaft.IsLeader(n) {
			RedirectToLeader(l, w, n)
			return
		}

		action, err := internalRaft.EncodeAction(l, internalRaft.StoreAction{
			Action: internalRaft.Delete,
			Key:    key,
		})
		if err != nil {
			internalError(l, r, w, err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := n.Propose(ctx, action); err != nil {
			internalError(l, r, w, err)
			return
		}
	})
}

// @title Status
// @description gets raft state
// @success 200
// @router /status/ [get]
func NewStatusHandler(l *slog.Logger, n raft.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, err := n.Status().MarshalJSON()
		if err != nil {
			internalError(l, r, w, err)
			return
		}

		w.Write(status)
	})
}
