package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Matvey-Makaro/url-shortener/internal/lib/api/response"
	"github.com/Matvey-Makaro/url-shortener/internal/lib/logger/sl"
	"github.com/Matvey-Makaro/url-shortener/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirect.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		log.Info("alias", slog.String("alias", alias))
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		url, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", sl.Err(err))
			render.JSON(w, r, response.Error("url not found"))
			return
		}
		if err != nil {
			log.Info("get url error", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("got url", slog.String("url", url))

		http.Redirect(w, r, url, http.StatusFound)
	}
}
