package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrResponse struct {
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// bodyをJSON にしてHTTPレスポンスとして返す
func RespondJSON(ctx context.Context, w http.ResponseWriter, body any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// bodyをJSON に変換
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("encode response error: %v", err)

		// ステータスを500に設定する
		w.WriteHeader(http.StatusInternalServerError)

		// エラーレスポンス作成
		rsp := ErrResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		}

		// JSONで書き込む
		err := json.NewEncoder(w).Encode(rsp)
		if err != nil {
			fmt.Printf("write error response error: %v", err)
		}
		return
	}

	w.WriteHeader(status)

	// JSONを書き込む
	// fmt.Fprintf(w, "%s", bodyBytes) より早い
	_, err = w.Write(bodyBytes)
	if err != nil {
		fmt.Printf("write response error: %v", err)
	}
}

// w.Write(bodyBytes) : すでにJSONになっているデータをそのまま書き込む

// json.NewEncoder(w).Encode(rsp) : JSON変換 + 書き込みを同時に行う

// | 項目     | Write      | Encoder     |
// | ------ | ---------- | ----------- |
// | JSON変換 | 自分でmarshal | 自動          |
// | 書き込み   | Writeのみ    | Encodeが両方やる |
// | 改行     | なし         | 最後に`\n`     |
// | エラー処理  | 2回必要       | 1回          |
