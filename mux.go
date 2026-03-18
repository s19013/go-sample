package main

import "net/http"

func NewMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json; charaset=utf-8")

		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}
