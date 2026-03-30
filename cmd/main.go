package main

import (
	"fmt"
	"log"
	"os"

	githttp "github.com/qoppa-tech/toy-gitfed/internal/api/http"
)

var DefaultPort = 8080

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <repos-dir>\n", os.Args[0])
		os.Exit(1)
	}

	reposDir := os.Args[1]

	addr := fmt.Sprintf("0.0.0.0:%d", DefaultPort)
	srv := githttp.NewServer(githttp.Config{
		ReposDir: reposDir,
		Address:  addr,
	})

	log.Fatal(srv.Serve())
}
