package main

import (
	"context"
	"log"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/runtime/bootstrap"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := bootstrap.DIDWWIngress(ctx, cfg); err != nil {
		log.Fatal(err)
	}
	log.Print("platform ingress bootstrap complete")
}
