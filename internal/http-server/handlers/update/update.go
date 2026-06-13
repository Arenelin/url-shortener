package update

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

type Request struct {
	Id   int64        `json:"id" validate:"required"`
	Task storage.Task `json:"task" validate:"required"`
}

type Response struct {
	resp.Response
	Task storage.TaskStructure `json:"task,required"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.4 --name=URLDelete
type TaskUpdate interface {
	UpdateTask(id int64, task storage.Task) (storage.TaskStructure, error)
}

func New(log *slog.Logger, taskUpdate TaskUpdate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.update.New"
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

		// 1. Проверяем, передал ли клиент имя задачи
		if req.Task.TaskName != nil {
			newName := *req.Task.TaskName
			// Вызываем обновление имени в БД
			log.Info("Обновляем имя задачи", slog.String("name", newName))
			// К примеру: err = taskUpdate.UpdateTaskName(alias, newName)
		}

		// 2. Проверяем, передал ли клиент статус задачи
		if req.Task.IsDone != nil {
			newStatus := *req.Task.IsDone
			// Вызываем обновление статуса в БД
			log.Info("Обновляем статус задачи", slog.Bool("status", newStatus))
			// К примеру: err = taskUpdate.UpdateTaskStatus(alias, newStatus)
		}

		updatedTask, err := taskUpdate.UpdateTask(req.Id, req.Task)

		if err != nil {
			log.Info("failed to update task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))
			// TODO: render.Status(r, http.StatusBadRequest)

			return
		}
		log.Info("task updated")

		responseOK(w, r, updatedTask)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, updatedTask storage.TaskStructure) {
	render.JSON(w, r, Response{
		Response: resp.OK(), Task: updatedTask,
	})
	render.Status(r, http.StatusCreated)
}
