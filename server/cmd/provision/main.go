package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/runtime/provisioning"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "didww" || os.Args[2] != "ingress" {
		fmt.Fprintln(os.Stderr, "usage: provision didww ingress")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := provisioning.ProvisionDIDWWIngress(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
	log.Print("DIDWW platform ingress provisioned")
}
