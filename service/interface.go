package service

import (
	"context"

	"github.com/s19013/go-sample/entity"
	"github.com/s19013/go-sample/store"
)

// storeパッケージの直接参照を避けるためのインターフェイス

//go:generate go run github.com/matryer/moq -out moq_test.go . TaskAdder TaskLister
type TaskAdder interface {
	AddTask(ctx context.Context, db store.Execer, t *entity.Task) error
}
type TaskLister interface {
	ListTasks(ctx context.Context, db store.Queryer) (entity.Tasks, error)
}
