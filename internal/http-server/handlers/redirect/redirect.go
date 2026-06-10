package redirect

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
)

type Response struct {
	resp.Response
	Tasks []string `json:"tasks"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.4 --name=URLGetter
type TasksGetter interface {
	GetAllTasks() ([]string, error)
}

func New(log *slog.Logger, tasksGetter TasksGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.get.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		tasks, err := tasksGetter.GetAllTasks()

		//if errors.Is(err, storage.ErrURLNotFound) {
		//	log.Info("url not found", slog.String("alias", alias))
		//
		//	render.JSON(w, r, resp.Error("url not found"))
		//
		//	return
		//}

		if err != nil {
			log.Info("failed to get tasks", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("got tasks", slog.String("tasks", string(rune(len(tasks)))))
		responseOK(w, r, tasks)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, tasks []string) {
	render.JSON(w, r, Response{
		Response: resp.OK(), Tasks: tasks,
	})
	render.Status(r, http.StatusCreated)
}
