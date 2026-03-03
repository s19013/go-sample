package main

import (
	"context"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Printf("need port number\n")
		os.Exit(1)
	}

	p := os.Args[1]
	l, err := net.Listen("tcp", ":"+p)
	if err != nil {
		log.Fatalf("failed to listen port %s : %v", p, err)
	}

	err = run(context.Background(), l)
	if err != nil {
		log.Printf("failed to terminate server: %v", err)
		os.Exit(1)
	}
}
