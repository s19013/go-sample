package main

import (
	"context"
	"log"
)

func main() {

	err := run(context.Background())
	if err != nil {
		log.Printf("failed to terminate server: %v", err)
	}
}
