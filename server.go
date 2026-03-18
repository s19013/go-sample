package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv *http.Server
	l   net.Listener
}

func (s *Server) Run(ctx context.Context) error {
	// Ctrl+C や SIGTERM を受け取ったときに処理を終了できるようにする Context
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

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
		log.Printf("listning:%v", s.l.Addr().String())

		err := s.srv.Serve(s.l)
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
	err := s.srv.Shutdown(context.Background())

	if err != nil {
		log.Printf("failed to terminate server: %v", err)
	}

	// Goメソッドで起動した goroutine 全部が終了するまで待つ
	// グレースフルシャットダウンの終了を待つ。
	return eg.Wait()
}

func NewServer(l net.Listener, mux http.Handler) *Server {
	return &Server{
		srv: &http.Server{Handler: mux},
		l:   l,
	}
}
