package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/s19013/go-sample/config"
	"github.com/s19013/go-sample/entity"
)

// KVS ->（Key-Value Store）

func NewKVS(ctx context.Context, cfg *config.Config) (*KVS, error) {
	// 設定（cfg.RedisHost, cfg.RedisPort）からRedisの接続先を作る
	// Redisクライアントを生成
	cli := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})

	// Pingで接続確認
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &KVS{Cli: cli}, nil
}

// Redisクライアントを持つだけのラッパー
// この struct を通して保存・取得する
type KVS struct {
	Cli *redis.Client
}

func (k *KVS) Save(ctx context.Context, key string, userID entity.UserID) error {
	// key に対して userID を保存
	// 30分の有効期限付き（TTL）
	id := int64(userID)
	return k.Cli.Set(ctx, key, id, 30*time.Minute).Err()
}

func (k *KVS) Load(ctx context.Context, key string) (entity.UserID, error) {
	// Redisから値を取得
	// int64として読み取る

	id, err := k.Cli.Get(ctx, key).Int64()

	// 取得失敗（存在しない・期限切れ）ならエラー
	if err != nil {
		return 0, fmt.Errorf("failed to get by %q: %w", key, ErrNotFound)
	}

	// int64 → entity.UserID に型変換して返している
	// 「これはユーザーIDですよ」と明示できる

	// id情報しか保存してないためid情報しか返せない
	return entity.UserID(id), nil
}
