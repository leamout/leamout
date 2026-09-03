package main

import (
	"context"
	"os"

	"github.com/leamout/leamout/internal/operatorcli"
)

var (
	version = "dev"
	commit  = "unknown"
	builtAt = "unknown"
)

func main() {
	os.Exit(operatorcli.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, operatorcli.BuildInfo{
		Version: version,
		Commit:  commit,
		BuiltAt: builtAt,
	}))
}
