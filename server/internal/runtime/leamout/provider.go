package leamout

import (
	"context"
	"io"
)

func runProvider(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	if len(args) != 2 || args[0] != "didww" || args[1] != "provision-ingress" {
		writeln(stderr, "usage: leamout provider didww provision-ingress")
		return 2
	}

	code := runInstalledCompose(ctx, stdout, stderr,
		"run", "--rm", "--no-deps", "server", "/leamout/provision", "didww", "ingress",
	)
	if code == 0 {
		writeln(stdout, "✓ DIDWW platform ingress provisioned")
	}
	return code
}
