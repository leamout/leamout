package leamout

import (
	"context"
	"io"
	"path/filepath"
)

func runUpdate(ctx context.Context, stdout, stderr io.Writer, args []string, version string) int {
	if len(args) != 0 {
		writeln(stderr, "usage: leamout update")
		return 2
	}
	if version == "" || version == "dev" {
		writeln(stderr, "development CLI cannot install a production update")
		return 1
	}
	state, err := loadDeploymentState("/var/lib/leamout/deployment.json")
	if err != nil {
		writef(stderr, "load deployment identity: %v\n", err)
		return 1
	}
	if err := validateRuntimeEnv("/etc/leamout/leamout.env", state.DeploymentID); err != nil {
		writef(stderr, "validate deployment configuration: %v\n", err)
		return 1
	}
	if err := installRuntimeBundle(filepath.Join("/var/lib/leamout/releases", version), "/var/lib/leamout/runtime", version); err != nil {
		writef(stderr, "install staged runtime: %v\n", err)
		return 1
	}
	if code := runInstalledCompose(ctx, stdout, stderr, "pull"); code != 0 {
		return code
	}
	if code := runInstalledCompose(ctx, stdout, stderr, "up", "-d", "--remove-orphans"); code != 0 {
		return code
	}
	writeln(stdout, "✓ Runtime update installed")
	writeln(stdout, "Run: sudo leamout doctor")
	return 0
}
