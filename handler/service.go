package handler

import (
	"context"

	"github.com/s19013/go-sample/entity"
)

// リクエストの解釈、レスポンスを組み立てる処理以外をこちらに委譲する

//go:generate go run github.com/matryer/moq -out moq_test.go . ListTasksService AddTaskService
type ListTasksService interface {
	ListTasks(ctx context.Context) (entity.Tasks, error)
}
type AddTaskService interface {
	AddTask(ctx context.Context, title string) (*entity.Task, error)
}
