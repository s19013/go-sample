package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/s19013/go-sample/entity"
	"github.com/s19013/go-sample/testutil"
)

func TestAddTask(t *testing.T) {
	type want struct {
		status  int
		rspFile string
	}

	// テストケース2を定義
	tests := map[string]struct {
		reqFile string
		want    want
	}{
		"ok": {
			reqFile: "testdata/add_task/ok_req.json.golden",
			want: want{
				status:  http.StatusOK,
				rspFile: "testdata/add_task/ok_rsp.json.golden",
			},
		},
		"badRequest": {
			reqFile: "testdata/add_task/bad_req.json.golden",
			want: want{
				status:  http.StatusBadRequest,
				rspFile: "testdata/add_task/bad_rsp.json.golden",
			},
		},
	}

	// n → "ok" や "badRequest"（キー）
	// tt → 各テストケースの中身（struct）
	for n, tt := range tests {
		// Goの for range はクセがあって
		// ループ変数は1個しか使い回される
		// 全部「最後のtt」になる可能性あり（並列だと特に）
		// その時点のttを新しい変数にコピーして各テストが正しい値を持つ
		tt := tt

		// サブテストを作成
		// 今回の場合、ok,badRequestでテストを分割できる
		t.Run(n, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodPost,
				"/tasks",
				bytes.NewReader(testutil.LoadFile(t, tt.reqFile)),
			)

			// サービスのモック作成
			moq := &AddTaskServiceMock{}

			// モックを定義
			moq.AddTaskFunc = func(
				ctx context.Context, title string,
			) (*entity.Task, error) {
				// 成功ケース
				if tt.want.status == http.StatusOK {
					return &entity.Task{ID: 1}, nil
				}

				// 失敗ケース
				return nil, errors.New("error from mock")
			}

			// テスト実行
			// AddTaskを作る
			sut := AddTask{
				Service:   moq,
				Validator: validator.New(),
			}

			sut.ServeHTTP(w, r)

			resp := w.Result()

			testutil.AssertResponse(
				t,
				resp,
				tt.want.status,
				testutil.LoadFile(t, tt.want.rspFile),
			)
		})
	}
}
