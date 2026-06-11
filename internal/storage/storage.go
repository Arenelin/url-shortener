package storage

import "errors"

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTaskExists   = errors.New("task exists")
)

type Task struct {
	TaskName *string `json:"taskName" validate:"omitempty,required"`
	IsDone   *bool   `json:"isDone" validate:"omitempty"`
}

type TaskStructure struct {
	TaskName *string `json:"taskName" validate:"omitempty,required"`
	IsDone   *bool   `json:"isDone" validate:"omitempty"`
	Id       *int64  `json:"id" validate:"omitempty"`
}
