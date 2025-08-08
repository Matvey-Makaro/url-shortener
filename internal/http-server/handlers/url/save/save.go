package save

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
	_ "github.com/vektra/mockery"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty" validate:"required"`
}

type Response struct {
	response.Response
	Alias string `json:"alias"`
}

type URLSaver interface {
	SaveURL(urlToSave, alias string) error
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

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

		// TODO: Можно добавить логику, что если alias не передан, то генерировать рандомный
		err = urlSaver.SaveURL(req.URL, req.Alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("alias already exists", slog.String("alias", req.Alias))
			render.JSON(w, r, response.Error("alias already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, response.Error("failed to add url"))
			return
		}

		log.Info("url added")
		responseOK(w, r, req.Alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Alias:    alias,
	})
}
