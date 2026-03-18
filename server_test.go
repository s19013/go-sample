package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestServerRun(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("failed to listen port  %v", err)
	}

	// キャンセル可能な｢context.Context｣のオブジェクトを作る
	ctx, cancel := context.WithCancel(context.Background())

	// 別ゴルーチンでテスト対象の｢run｣を実行してHttpサーバーを起動
	eg, ctx := errgroup.WithContext(ctx)
	mux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, %s", r.URL.Path[1:])
		if err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	eg.Go(func() error {
		s := NewServer(l, mux)
		return s.Run(ctx)
	})

	// リクエスト送信
	in := "message"
	url := fmt.Sprintf("http://%s/%s", l.Addr().String(), in)
	rsp, err := http.Get(url)
	if err != nil {
		t.Errorf("failed to terminate server: %v", err)
	}

	defer rsp.Body.Close()
	got, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	// httpサーバーの戻り値を検証する
	want := fmt.Sprintf("Hello, %s", in)
	if string(got) != want {
		t.Errorf("want %q,but got %q", want, got)
	}

	// run関数に終了通知を送信
	cancel()

	// run関数の戻り値を検証する
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
