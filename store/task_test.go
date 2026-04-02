package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/s19013/go-sample/clock"
	"github.com/s19013/go-sample/entity"
	"github.com/s19013/go-sample/testutil"
	"github.com/s19013/go-sample/testutil/fixture"
)

// ユーザーを1件DBに登録してIDを返す
func prepareUser(ctx context.Context, t *testing.T, db Execer) entity.UserID {
	t.Helper()

	u := fixture.User(nil)

	result, err := db.ExecContext(
		ctx,
		"INSERT INTO user (name, password, role, created, modified) VALUES (?, ?, ?, ?, ?);",
		u.Name, u.Password, u.Role, u.Created, u.Modified,
	)

	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("got user_id: %v", err)
	}

	return entity.UserID(id)
}

// タスクを複数登録して、「テストで期待するデータ」を返す
func prepareTasks(ctx context.Context, t *testing.T, con Execer) (entity.UserID, entity.Tasks) {
	t.Helper()

	// ユーザーデータを作成
	userID := prepareUser(ctx, t, con)
	otherUserID := prepareUser(ctx, t, con)

	// 固定時刻を使って期待値を作る
	c := clock.FixedClocker{}

	wants := entity.Tasks{
		{
			UserID: userID,
			Title:  "want task 1", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		{
			UserID: userID,
			Title:  "want task 2", Status: "done",
			Created: c.Now(), Modified: c.Now(),
		},
	}

	// ログインユーザーだけを取得できるか試すため、「他人のタスク」を混ぜてる
	tasks := entity.Tasks{
		wants[0],
		wants[1],
		{
			UserID: otherUserID,
			Title:  "not want task", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
	}

	// まとめてINSERT
	result, err := con.ExecContext(ctx,
		`INSERT INTO task (user_id, title, status, created, modified)
			VALUES
			    (?, ?, ?, ?, ?),
			    (?, ?, ?, ?, ?),
			    (?, ?, ?, ?, ?);`,
		tasks[0].UserID, tasks[0].Title, tasks[0].Status, tasks[0].Created, tasks[0].Modified,
		tasks[1].UserID, tasks[1].Title, tasks[1].Status, tasks[1].Created, tasks[1].Modified,
		tasks[2].UserID, tasks[2].Title, tasks[2].Status, tasks[2].Created, tasks[2].Modified,
	)

	if err != nil {
		t.Fatal(err)
	}

	// 採番されたIDを期待値に入れる
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	// INSERTしただけじゃIDは struct に入らない
	// 連番で入る前提でIDを補完
	// 期待値と一致させるため
	tasks[0].ID = entity.TaskID(id)
	tasks[1].ID = entity.TaskID(id + 1)
	tasks[2].ID = entity.TaskID(id + 2)
	return userID, wants
}

func TestRepository_ListTasks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// entity.Taskを作成する他のテストケースと混ざるとテストがフェイルする。
	// そのため、トランザクションをはることでこのテストケースの中だけのテーブル状態にする。
	tx, err := testutil.OpenDBForTest(t).BeginTxx(ctx, nil)

	// このテストケースが完了したらもとに戻す
	t.Cleanup(func() { _ = tx.Rollback() })
	if err != nil {
		t.Fatal(err)
	}

	wantUserID, wants := prepareTasks(ctx, t, tx)

	// act
	sut := &Repository{}
	gots, err := sut.ListTasks(ctx, tx, wantUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// assert
	if d := cmp.Diff(gots, wants); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s", d)
	}
}

// sqlmock を使って AddTask のSQL実行を検証する
func TestRepository_AddTask(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	c := clock.FixedClocker{}

	// 期待するタスクを作る
	var wantID int64 = 20
	okTask := &entity.Task{
		UserID:   33,
		Title:    "ok task",
		Status:   "todo",
		Created:  c.Now(),
		Modified: c.Now(),
	}

	// モックDBを作るs
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// 「AddTask を呼んだら、次のことが起きるはず」と定義しています。
	// INSERT INTO task ... が実行される
	// 引数は Title, Status, Created, Modified
	// 実行結果として
	// LastInsertId = 20
	// RowsAffected = 1
	mock.ExpectExec(
		// エスケープが必要
		`INSERT INTO task \(user_id, title, status, created, modified\) VALUES \(\?, \?, \?, \?, \?\)`,
	).WithArgs(
		okTask.UserID, okTask.Title, okTask.Status, okTask.Created, okTask.Modified,
	).WillReturnResult(sqlmock.NewResult(wantID, 1))

	xdb := sqlx.NewDb(db, "mysql")

	// act
	r := &Repository{Clocker: c}

	// assert
	if err := r.AddTask(ctx, xdb, okTask); err != nil {
		t.Errorf("want no error, but got %v", err)
	}
}
