package main

import (
	"context"
	"os"

	"github.com/leamout/leamout/internal/runtime/leamout"
)

var (
	version = "dev"
	commit  = "unknown"
	builtAt = "unknown"
)

func main() {
	os.Exit(leamout.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, leamout.BuildInfo{
		Version: version,
		Commit:  commit,
		BuiltAt: builtAt,
	}))
}
