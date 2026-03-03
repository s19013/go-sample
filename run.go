package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context) error {
	// HTTPサーバーの定義
	s := &http.Server{
		Addr: ":18080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "hello, %s", r.URL.Path[1:])
		}),
	}

	// errgroup の作成
	// * goroutineのエラーをまとめて管理する仕組み
	// * どれか1つがエラーになったら全体をキャンセルできる
	// * ctx はキャンセル可能なContextになる

	// 複数のgoroutineを安全に管理するための準備。

	eg, ctx := errgroup.WithContext(ctx)

	// 別ゴルーチンでhttpサーバーを起動
	// メイン処理をブロックしないため

	eg.Go(func() error {

		err := s.ListenAndServe()
		if err != nil {

			// http.ErrServerCloseは
			// http.Server.ShutDown()が正常に終了したことを示すので何もしない
			if err == http.ErrServerClosed {
				return nil
			}

			log.Printf("failed to terminate server: %v", err)
			return err
		}

		return nil
	})

	// キャンセルされるまでブロック（待機）する

	// * ctx は context.Context
	// * Done() は「キャンセル通知用のチャネル」
	// * <- は「チャネル受信」

	<-ctx.Done()

	// HTTPサーバーを安全に停止させる
	err := s.Shutdown(context.Background())

	if err != nil {
		log.Printf("failed to terminate server: %v", err)
	}

	// Goメソッドで起動した goroutine 全部が終了するまで待つ
	return eg.Wait()
}
