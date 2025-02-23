package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	internalRaft "github.com/pablovarg/distributed-key-value-store/raft"
	"go.etcd.io/raft/v3"
)

// @title Put
// @description inserts or updates a key's value
// @accept json
// @param input body api.NewPutHandler.input true "Key / Value pair"
// @success 200
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
		n.Propose(ctx, a)
	})
}
