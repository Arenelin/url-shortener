package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/cmd/internal/lib/api/response"
	"url-shortener/cmd/internal/lib/logger/sl"
	"url-shortener/cmd/internal/storage"
)

type URLDelete interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDelete URLDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		log.Info("alias successfully received", slog.String("alias", alias))

		err := urlDelete.DeleteURL(alias)

		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", alias))

			render.JSON(w, r, resp.Error("url not found"))

			return
		}

		if err != nil {
			log.Info("failed to delete url", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))
			// TODO: render.Status(r, http.StatusBadRequest)

			return
		}
		log.Info("URL deleted")
	}
}
