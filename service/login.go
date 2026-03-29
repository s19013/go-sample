package service

import (
	"context"
	"fmt"

	"github.com/s19013/go-sample/store"
)

type Login struct {
	DB             store.Queryer
	Repo           UserGetter
	TokenGenerator TokenGenerator
}

// ログイン情報の検証とアクセストークンの生成を行う
func (l *Login) Login(ctx context.Context, name, pw string) (string, error) {
	// ユーザー情報取得
	u, err := l.Repo.GetUser(ctx, l.DB, name)
	if err != nil {
		return "", fmt.Errorf("failed to list: %w", err)
	}

	// パスワード検証
	if err := u.ComparePassword(pw); err != nil {
		return "", fmt.Errorf("wrong password: %w", err)
	}

	// トークン生成
	jwt, err := l.TokenGenerator.GenerateToken(ctx, *u)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return string(jwt), nil
}
