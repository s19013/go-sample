package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/s19013/go-sample/config"
	"golang.org/x/sync/errgroup"
)

func main() {
	err := run(context.Background())
	if err != nil {
		log.Printf("failed to terminated server: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen port %d: %v", cfg.Port, err)
	}
	url := fmt.Sprintf("http://%s", l.Addr().String())
	log.Printf("start with: %v", url)

	// HTTPサーバーの定義
	s := &http.Server{
		// 引数で受け取ったnet.listenerを利用するので
		// addrフィールドは指定しない
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, "hello, %s", r.URL.Path[1:])
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
		// ポート表示
		log.Printf("listning:%v", l.Addr().String())

		// addrフィールドを外から決めるため、serveに変更
		err := s.Serve(l)
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
	err = s.Shutdown(context.Background())

	if err != nil {
		log.Printf("failed to terminate server: %v", err)
	}

	// Goメソッドで起動した goroutine 全部が終了するまで待つ
	return eg.Wait()
}
