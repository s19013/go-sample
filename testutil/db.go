package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// テスト用にDB接続を作って *sqlx.DB を返す
func OpenDBForTest(t *testing.T) *sqlx.DB {
	// 境に応じてポートを変える
	// ローカル → 33306
	// CI環境（GitHub Actionsとか） → 3306
	port := 33306
	if _, defined := os.LookupEnv("CI"); defined {
		port = 3306
	}

	// DB接続作成
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf("todo:todo@tcp(127.0.0.1:%d)/todo?parseTime=true", port),
	)
	if err != nil {
		t.Fatal(err)
	}

	// テスト終了時にDB接続を自動で閉じる
	t.Cleanup(
		func() { _ = db.Close() },
	)
	return sqlx.NewDb(db, "mysql")
}
