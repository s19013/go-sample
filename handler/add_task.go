package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/s19013/go-sample/entity"
)

type AddTask struct {
	Service   AddTaskService
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
	if err := validator.New().Struct(b); err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// Task作成
	t, err := at.Service.AddTask(ctx, b.Title)

	if err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	rsp := struct {
		ID entity.TaskID `json:"id"`
	}{ID: t.ID} // ここでstructに値を入れてる

	RespondJSON(ctx, w, rsp, http.StatusOK)
}
