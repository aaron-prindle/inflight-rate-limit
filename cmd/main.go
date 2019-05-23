package main

import (
	"log"
	"net/http"

	"fmt"

	"github.com/yue9944882/inflight-rate-limit"
)

func main() {

	queues := inflight.InitQueues(1)
	fqFilter := inflight.NewFQFilter(queues)

	fqFilter.Run()

	http.HandleFunc("/", fqFilter.Serve)
	fmt.Printf("Serving at 127.0.0.1:8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
