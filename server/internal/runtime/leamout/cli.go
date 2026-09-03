package leamout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	selfHostedBuildServices  = "server worker opensips freeswitch rtpengine"
	selfHostedDeployServices = "server worker opensips coturn freeswitch rtpengine postgres redis nats"
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
		return runInit(stdout, stderr)
	case "up":
		return runScript(ctx, stdout, stderr, ".", ".env", "deploy/compose.yaml", "deploy/certs", "server/scripts/deploy/up.sh")
	case "down":
		return runCompose(ctx, stdout, stderr, ".", ".env", "deploy/compose.yaml", "down")
	case "status":
		return runCompose(ctx, stdout, stderr, ".", ".env", "deploy/compose.yaml", "ps")
	case "logs":
		composeArgs := []string{"logs", "-f", "--tail=200"}
		composeArgs = append(composeArgs, args[1:]...)
		return runCompose(ctx, stdout, stderr, ".", ".env", "deploy/compose.yaml", composeArgs...)
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

	for _, name := range []string{"docker", "sh", "curl"} {
		if _, err := exec.LookPath(name); err != nil {
			writef(stderr, "missing prerequisite: %s\n", name)
			return 1
		}
	}
	writeln(stdout, "✓ linux/amd64 host")
	writeln(stdout, "✓ docker, sh, and curl available")

	if code := runScript(ctx, stdout, stderr, ".", ".env", "deploy/compose.yaml", "deploy/certs", "server/scripts/deploy/preflight.sh"); code != 0 {
		return code
	}
	writeln(stdout, "Leamout doctor passed.")
	return 0
}

func runScript(ctx context.Context, stdout, stderr io.Writer, rootPath, envFile, composeFile, certDir, rel string) int {
	root, err := filepath.Abs(rootPath)
	if err != nil {
		writef(stderr, "resolve Leamout root: %v\n", err)
		return 1
	}
	path := filepath.Join(root, rel)
	if _, err := os.Stat(path); err != nil {
		writef(stderr, "required deployment primitive missing: %s\n", path)
		return 1
	}

	cmd := exec.CommandContext(ctx, "sh", path)
	cmd.Dir = root
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(),
		"ENV_FILE="+envFile,
		"COMPOSE_FILE="+composeFile,
		"CERT_DIR="+certDir,
		"BUILD_SERVICES="+selfHostedBuildServices,
		"DEPLOY_SERVICES="+selfHostedDeployServices,
	)
	return exitCode(cmd.Run(), stderr)
}

func runCompose(ctx context.Context, stdout, stderr io.Writer, rootPath, envFile, composeFile string, args ...string) int {
	root, err := filepath.Abs(rootPath)
	if err != nil {
		writef(stderr, "resolve Leamout root: %v\n", err)
		return 1
	}
	base := []string{"compose", "--env-file", envFile, "-f", composeFile}
	base = append(base, args...)
	cmd := exec.CommandContext(ctx, "docker", base...)
	cmd.Dir = root
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
  init       Initialize durable self-hosted deployment identity, secrets, and configuration
  up         Start self-hosted runtime services through existing deployment primitives
  down       Stop the existing Leamout deployment stack
  status     Show deployment service status
  logs       Follow deployment logs
  doctor     Validate host prerequisites and deployment preflight
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
