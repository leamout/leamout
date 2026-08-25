package main

import (
	"log"

	"github.com/leamout/leamout/server/internal/runtime/worker"
)

func main() {
	w := worker.New()
	if err := w.Start(); err != nil {
		log.Fatal(err)
	}
}
