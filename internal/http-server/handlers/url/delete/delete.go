package delete

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Matvey-Makaro/url-shortener/internal/lib/api/response"
	"github.com/Matvey-Makaro/url-shortener/internal/lib/logger/sl"
	"github.com/Matvey-Makaro/url-shortener/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Alias string `json:"alias" validate:"required"`
}

type Response struct {
	response.Response
}

type AliasDeleter interface {
	DeleteAlias(alias string) error
}

func New(log *slog.Logger, aliasDeleter AliasDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("empty request")
			render.JSON(w, r, response.Error("empty request"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("req", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request:", sl.Err(validateErr))
			render.JSON(w, r, response.ValidationErrors(validateErr))
			return
		}

		err = aliasDeleter.DeleteAlias(req.Alias)
		if errors.Is(err, storage.ErrAliasNotFound) {
			log.Info("alias doesn't exist")
			render.JSON(w, r, response.Error("alias doesn't exist"))
			return
		}
		if err != nil {
			log.Error("failed to delete alias", sl.Err(err))
			render.JSON(w, r, response.Error("failed to delete alias"))
			return
		}
		log.Info("alias deleted", slog.String("alias", req.Alias))
		render.JSON(w, r, response.OK())
	}
}
