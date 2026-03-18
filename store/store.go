package store

import (
	"errors"

	"github.com/s19013/go-sample/entity"
)

var (
	Tasks = &TaskStore{Tasks: map[entity.TaskID]*entity.Task{}}

	ErrNotFound = errors.New("not found")
)

type TaskStore struct {
	// 動作確認用の仮実装なのであえてexportしている。
	LastID entity.TaskID

	// TaskIDをキーにしてTaskへのポインタを保存する連想配列
	Tasks map[entity.TaskID]*entity.Task
}

func (ts *TaskStore) Add(t *entity.Task) (entity.TaskID, error) {
	ts.LastID++

	// TaskにIDをセット
	t.ID = ts.LastID

	// mapに保存
	ts.Tasks[t.ID] = t

	return t.ID, nil
}

// ソート済みのタスク一覧を返す
func (ts *TaskStore) All() entity.Tasks {
	// make は スライス / map / channel を作るための組み込み関数
	// make(型, 長さ, 容量)

	// Tasksの要素数と同じ長さの Taskポインタスライスを作る
	tasks := make([]*entity.Task, 0, len(ts.Tasks))
	for _, t := range ts.Tasks {
		tasks = append(tasks, t)
	}
	return tasks
}
