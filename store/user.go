package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/s19013/go-sample/entity"
)

func (r *Repository) RegisterUser(ctx context.Context, db Execer, u *entity.User) error {
	u.Created = r.Clocker.Now()
	u.Modified = r.Clocker.Now()

	sql := `INSERT INTO user (name, password, role, created, modified) VALUES (?, ?, ?, ?, ?)`

	result, err := db.ExecContext(ctx, sql, u.Name, u.Password, u.Role, u.Created, u.Modified)
	if err != nil {
		// MySQLエラーかどうか判定
		// errors.As は「このエラー、指定した型に変換できる？」というチェック
		// エラー番号で種類を判定
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) &&
			mysqlErr.Number == ErrCodeMySQLDuplicateEntry {
			// “重複エラー”なら → 独自エラーに変換
			return fmt.Errorf("cannot create same name user: %w", ErrAlreadyEntry)
		}
		// それ以外はそのまま返す
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	u.ID = entity.UserID(id)
	return nil
}

func (r *Repository) GetUser(
	ctx context.Context, db Queryer, name string,
) (*entity.User, error) {
	u := &entity.User{}
	sql := `SELECT
		id, name, password, role, created, modified 
		FROM user WHERE name = ?`
	if err := db.GetContext(ctx, u, sql, name); err != nil {
		return nil, err
	}
	return u, nil
}
