package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/s19013/go-sample/entity"
	"github.com/s19013/go-sample/store"
)

type AddTask struct {
	Store     *store.TaskStore
	Validator *validator.Validate
}

// http.Handlerインターフェース実装
// http.Handle("/tasks", &AddTask{}) のように使える
func (at *AddTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// このリクエストの状態を管理するオブジェクトを取得
	ctx := r.Context()

	// 無名構造体を作って、JSONを受け取るための入れ物にしている
	var b struct {
		Title string `json:"title" validate:"required"`
	}

	// デコード
	// JSONを 上記で定義している b に変換する
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// バリデーション
	err := validator.New().Struct(b)
	if err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// Task作成
	t := &entity.Task{
		Title:   b.Title,
		Status:  "todo",
		Created: time.Now(),
	}

	// ストアに保存
	id, err := at.Store.Add(t)
	if err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	rsp := struct {
		ID entity.TaskID `json:"id"`
	}{ID: id} // ここで値を入れてる

	RespondJSON(ctx, w, rsp, http.StatusOK)
}
