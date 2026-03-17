package config

// 環境変数を含む設定項目を管理する構造体を作成

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	// 環境変数（Environment Variables）から設定値を読み込んで Config 構造体に入れる処理
	// env:"環境変数名" envDefault:"デフォルト値"
	Env  string `env:"TODO_ENV" envDefault:"dev"`
	Port int    `env:"PORT" envDefault:"80"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// 空のConfigを作る
// 環境変数を読み込んで埋める
// Configを返す
