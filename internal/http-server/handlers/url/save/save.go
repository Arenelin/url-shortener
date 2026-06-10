package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"
)

type Request struct {
	TaskName string `json:"taskName" validate:"required"`
}

type Response struct {
	resp.Response
	TaskName string `json:"taskName,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.4 --name=URLSaver
type TaskSaver interface {
	CreateTask(taskName string) (int64, error)
}

func New(log *slog.Logger, taskSaver TaskSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.save.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			//log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			//log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		taskName := req.TaskName

		id, err := taskSaver.CreateTask(taskName)

		if errors.Is(err, storage.ErrTaskExists) {
			log.Info("task already exists", slog.String("task", req.TaskName))

			render.JSON(w, r, resp.Error("url already exists"))
			render.Status(r, http.StatusBadRequest)

			return
		}
		if err != nil {
			//log.Info("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error(err.Error()))
			render.Status(r, http.StatusBadRequest)

			return
		}
		log.Info("url added", slog.Int64("id", id))

		responseOK(w, r, taskName)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, taskName string) {
	render.JSON(w, r, Response{
		Response: resp.OK(), TaskName: taskName,
	})
	render.Status(r, http.StatusCreated)
}
