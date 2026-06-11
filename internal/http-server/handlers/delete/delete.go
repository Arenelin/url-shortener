package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.4 --name=URLDelete
type TaskDelete interface {
	DeleteTask(id int64) error
}

func New(log *slog.Logger, taskDelete TaskDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.delete.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		id := chi.URLParam(r, "id")

		if id == "" {
			log.Info("id is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Error("failed to parse id", slog.String("id", id))
			render.JSON(w, r, resp.Error("invalid id format"))
			return
		}

		log.Info("id successfully received", slog.Int64("id", idInt))

		err = taskDelete.DeleteTask(idInt)
		if errors.Is(err, storage.ErrTaskNotFound) {
			log.Info("task not found", slog.String("id", id))

			render.JSON(w, r, resp.Error("task not found"))

			return
		}

		if err != nil {
			log.Info("failed to delete task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))
			// TODO: render.Status(r, http.StatusBadRequest)

			return
		}
		log.Info("task deleted")

	}
}
