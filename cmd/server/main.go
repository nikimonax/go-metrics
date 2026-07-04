package main

import (
	"fmt"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/server"
)

func main() {
	server := server.New()

	err := http.ListenAndServe("localhost:8080", server)

	if err != nil {
		fmt.Println(err)
	}
}
