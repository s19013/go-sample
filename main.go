package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	err := http.ListenAndServe(
		":18080",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "hello, %s", r.URL.Path[1:])
		}),
	)
	if err != nil {
		fmt.Printf("failed to termiate server: %w", err)
		os.Exit(1)
	}
}
