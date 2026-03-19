package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// このパッケージの“init処理だけ使いたい”から _ で読み込んでる
	// コード内では一切使えない でも「読み込まれるだけ」

	// sql.openで"mysql"を使うため "mysql" が登録されてる必要がある
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/s19013/go-sample/clock"
	"github.com/s19013/go-sample/config"
)

// DB本体,Close関数（後で使う）,error を返す
func New(ctx context.Context, cfg *config.Config) (*sqlx.DB, func(), error) {
	// sqlx.Connectを使うと内部でpingする。

	// sql.Openは「接続しない」 ただの準備だけ
	db, err := sql.Open("mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.DBUser, cfg.DBPassword,
			cfg.DBHost, cfg.DBPort,
			cfg.DBName,
		),
	)

	if err != nil {
		return nil, func() {}, err
	}

	// Pingで接続確認
	// 実際にDBに繋がるかチェック
	// 2秒でタイムアウト
	// ここで初めて「接続確認」する
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, func() { _ = db.Close() }, err
	}

	// sql → sqlx に変換
	xdb := sqlx.NewDb(db, "mysql")

	return xdb, func() { _ = db.Close() }, nil
}

type Repository struct {
	// Clockerは「現在時刻を取得するための抽象」（テストしやすくするために使うやつ）
	Clocker clock.Clocker
}

// インターフェイスを使ってモックしやすくしている

// トランザクション
type Beginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// SQL準備
type Preparer interface {
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
}

// 書き込み
// INSERT / UPDATE / DELETE系
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// 読み取り
// SELECT系
type Queryer interface {
	Preparer
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...any) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...any) error
}

var (
	// インターフェイスが期待通りに宣言されているか確認
	_ Beginner = (*sqlx.DB)(nil)
	_ Preparer = (*sqlx.DB)(nil)
	_ Queryer  = (*sqlx.DB)(nil)
	_ Execer   = (*sqlx.DB)(nil)
	_ Execer   = (*sqlx.Tx)(nil)
)
