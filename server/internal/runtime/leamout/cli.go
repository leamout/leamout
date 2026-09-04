package leamout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type BuildInfo struct {
	Version string
	Commit  string
	BuiltAt string
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, build BuildInfo) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "--help", "-h":
		printHelp(stdout)
		return 0
	case "version", "--version", "-v":
		writef(stdout, "leamout %s\ncommit: %s\nbuilt: %s\n", build.Version, build.Commit, build.BuiltAt)
		return 0
	case "init":
		return runInit(stdout, stderr, build.Version)
	case "up":
		return runInstalledCompose(ctx, stdout, stderr, "up", "-d")
	case "down":
		return runInstalledCompose(ctx, stdout, stderr, "down")
	case "status":
		return runInstalledCompose(ctx, stdout, stderr, "ps")
	case "logs":
		composeArgs := []string{"logs", "-f", "--tail=200"}
		composeArgs = append(composeArgs, args[1:]...)
		return runInstalledCompose(ctx, stdout, stderr, composeArgs...)
	case "doctor":
		return runDoctor(ctx, stdout, stderr)
	default:
		writef(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return 2
	}
}

func runDoctor(ctx context.Context, stdout, stderr io.Writer) int {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		writef(stderr, "unsupported host: %s/%s; Self-Hosted Production v0.1 supports linux/amd64\n", runtime.GOOS, runtime.GOARCH)
		return 1
	}
	if _, err := exec.LookPath("docker"); err != nil {
		writeln(stderr, "Leamout runtime prerequisite is unavailable")
		return 1
	}

	state, err := loadDeploymentState("/var/lib/leamout/deployment.json")
	if err != nil {
		writef(stderr, "deployment identity: %v\n", err)
		return 1
	}
	if err := validateRuntimeEnv("/etc/leamout/leamout.env", state.DeploymentID); err != nil {
		writef(stderr, "deployment configuration: %v\n", err)
		return 1
	}
	if err := validateInstalledRuntimeFiles("/var/lib/leamout/runtime"); err != nil {
		writef(stderr, "production runtime: %v\n", err)
		return 1
	}

	cmd := exec.CommandContext(ctx, "docker", "compose", "--env-file", "/etc/leamout/leamout.env", "-f", "/var/lib/leamout/runtime/compose.yaml", "config", "--quiet")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if code := exitCode(cmd.Run(), stderr); code != 0 {
		return code
	}

	writeln(stdout, "✓ Supported host")
	writeln(stdout, "✓ Deployment identity and secrets valid")
	writeln(stdout, "✓ Production runtime installed")
	writeln(stdout, "✓ Runtime configuration valid")
	writeln(stdout, "Leamout doctor passed.")
	return 0
}

func runInstalledCompose(ctx context.Context, stdout, stderr io.Writer, args ...string) int {
	if err := validateInstalledRuntimeFiles("/var/lib/leamout/runtime"); err != nil {
		writef(stderr, "production runtime is not initialized: %v\nRun: sudo leamout init\n", err)
		return 1
	}
	if _, err := os.Stat("/etc/leamout/leamout.env"); err != nil {
		writef(stderr, "production configuration is not initialized: %v\nRun: sudo leamout init\n", err)
		return 1
	}

	base := []string{"compose", "--env-file", "/etc/leamout/leamout.env", "-f", "/var/lib/leamout/runtime/compose.yaml"}
	base = append(base, args...)
	cmd := exec.CommandContext(ctx, "docker", base...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	return exitCode(cmd.Run(), stderr)
}

func exitCode(err error, stderr io.Writer) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	if !strings.Contains(err.Error(), "signal: killed") {
		writef(stderr, "%v\n", err)
	}
	return 1
}

func printHelp(w io.Writer) {
	writeln(w, `Leamout self-hosted operator CLI

Usage:
  leamout <command>

Commands:
  init       Initialize deployment identity, secrets, configuration, and production runtime
  up         Start the installed Leamout runtime
  down       Stop the installed Leamout runtime
  status     Show installed runtime service status
  logs       Follow installed runtime logs
  doctor     Validate the local Leamout deployment
  version    Print CLI build information
  help       Show this help`)
}

// CLI output is best-effort. Commands return lifecycle/process failures; a closed
// output stream must not replace the underlying operator result with a write error.
func writef(w io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		return
	}
}

func writeln(w io.Writer, args ...any) {
	if _, err := fmt.Fprintln(w, args...); err != nil {
		return
	}
}
