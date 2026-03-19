package store

import (
	"context"

	"github.com/s19013/go-sample/entity"
)

func (r *Repository) AddTask(
	ctx context.Context, db Execer, t *entity.Task,
) error {
	t.Created = r.Clocker.Now()
	t.Modified = r.Clocker.Now()

	sql := `
		INSERT INTO task
		(title, status, created, modified)
		VALUES (?, ?, ?, ?)
	`

	// insert実行
	result, err := db.ExecContext(
		ctx, sql, t.Title, t.Status,
		t.Created, t.Modified,
	)

	if err != nil {
		return err
	}

	// DBが生成した自動採番IDを取得
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// 呼び出し元でもID使えるようにしてる (ポインタで渡されるからこんなことができる)
	t.ID = entity.TaskID(id)
	return nil
}

func (r *Repository) ListTasks(
	ctx context.Context, db Queryer,
) (entity.Tasks, error) {
	tasks := entity.Tasks{}

	sql := `
		SELECT id, title, status, created, modified
		FROM task;
	`

	err := db.SelectContext(ctx, &tasks, sql)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
