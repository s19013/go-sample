package handler

import (
	"context"

	"github.com/s19013/go-sample/entity"
	"github.com/s19013/go-sample/store"
)

// リクエストの解釈、レスポンスを組み立てる処理以外をこちらに委譲する

// ソースコードを自動生成するための記述
//
//go:generate go run github.com/matryer/moq -out moq_test.go . ListTasksService AddTaskService
type ListTasksService interface {
	ListTasks(ctx context.Context) (entity.Tasks, error)
}
type AddTaskService interface {
	AddTask(ctx context.Context, title string) (*entity.Task, error)
}

type UserRegister interface {
	RegisterUser(ctx context.Context, db store.Execer, u *entity.User) error
}
