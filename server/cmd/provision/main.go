package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/provisioning"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) == 4 && os.Args[1] == "managed-carrier" && os.Args[2] == "didww" && os.Args[3] == "ingress" {
		if err := provisioning.ProvisionDIDWWIngress(ctx, cfg); err != nil {
			log.Fatal(err)
		}
		log.Print("DIDWW managed-carrier ingress provisioned")
		return
	}

	if len(os.Args) >= 3 && os.Args[1] == "provider-operations" {
		switch os.Args[2] {
		case "list":
			state := ""
			limit := 0
			if len(os.Args) >= 4 {
				state = os.Args[3]
			}
			if len(os.Args) >= 5 {
				limit, err = strconv.Atoi(os.Args[4])
				if err != nil {
					log.Fatalf("invalid provider operation limit: %v", err)
				}
			}
			if len(os.Args) > 5 {
				usage()
			}
			operations, err := provisioning.ListProviderOperationDiagnostics(ctx, cfg, state, limit)
			if err != nil {
				log.Fatal(err)
			}
			writeJSON(operations)
			return

		case "get":
			if len(os.Args) != 4 {
				usage()
			}
			id, err := uuid.Parse(os.Args[3])
			if err != nil {
				log.Fatalf("invalid provider operation id: %v", err)
			}
			operation, err := provisioning.GetProviderOperationDiagnostic(ctx, cfg, id)
			if err != nil {
				log.Fatal(err)
			}
			writeJSON(operation)
			return

		case "reconcile":
			if len(os.Args) != 4 {
				usage()
			}
			id, err := uuid.Parse(os.Args[3])
			if err != nil {
				log.Fatalf("invalid provider operation id: %v", err)
			}
			operation, err := provisioning.ScheduleProviderOperationReconciliation(ctx, cfg, id)
			if err != nil {
				log.Fatal(err)
			}
			writeJSON(operation)
			return
		}
	}

	usage()
}

func writeJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  provision managed-carrier didww ingress
  provision provider-operations list [state] [limit]
  provision provider-operations get <operation-id>
  provision provider-operations reconcile <operation-id>`)
	os.Exit(2)
}
