package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// t.Helper() は 「この関数はテストの補助関数（ヘルパー）ですよ」と Go のテストフレームワークに伝えるもの。

// JSON同士が同じかを比較する関数
func AssertJSON(t *testing.T, want, got []byte) {
	t.Helper()

	// JSON を Go の値に変換
	var jw, jg any
	if err := json.Unmarshal(want, &jw); err != nil {
		t.Fatalf("cannot unmarshal want %q: %v", want, err)
	}
	if err := json.Unmarshal(got, &jg); err != nil {
		t.Fatalf("cannot unmarshal got %q: %v", got, err)
	}

	// 比較
	diff := cmp.Diff(jg, jw)
	if diff != "" {
		t.Errorf("got differs: (-got +want)\n%s", diff)
	}
}

// HTTPレスポンスを検証する関数
func AssertResponse(t *testing.T, got *http.Response, status int, body []byte) {
	t.Helper()
	t.Cleanup(func() { _ = got.Body.Close() })

	// bodyを読み取る
	gb, err := io.ReadAll(got.Body)
	if err != nil {
		t.Fatal(err)
	}

	// ステータスコード確認
	if got.StatusCode != status {
		t.Fatalf("want status %d, but got %d, body: %q", status, got.StatusCode, gb)
	}

	// bodyが空なら比較しない
	if len(gb) == 0 && len(body) == 0 {
		// 期待としても実体としてもレスポンスボディがないので
		// AssertJSONを呼ぶ必要はない。
		return
	}

	AssertJSON(t, body, gb)
}

// テスト用JSONをファイルから読む
func LoadFile(t *testing.T, path string) []byte {
	t.Helper()

	// テスト用 JSON をロード。
	bt, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read from %q: %v", path, err)
	}
	return bt
}
