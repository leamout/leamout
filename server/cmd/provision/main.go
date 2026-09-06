package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/provisioning"
)

func main() {
	if len(os.Args) != 4 || os.Args[1] != "managed-carrier" || os.Args[2] != "didww" || os.Args[3] != "ingress" {
		fmt.Fprintln(os.Stderr, "usage: provision managed-carrier didww ingress")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := provisioning.ProvisionDIDWWIngress(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
	log.Print("DIDWW managed-carrier ingress provisioned")
}
