package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMux(t *testing.T) {
	// レスポンス記録用のオブジェクト
	w := httptest.NewRecorder()

	// テスト用リクエスト作成
	r := httptest.NewRequest(http.MethodGet, "/health", nil)

	sut := NewMux()

	// ハンドラを実行
	sut.ServeHTTP(w, r)
	resp := w.Result()

	// テストが終わったらbodyをクローズしてメモリを開放
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Error("want status code 200, but", resp.StatusCode)
	}

	// レスポンスのBodyを全部読む
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	want := `{"status": "ok"}`
	if string(got) != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}
