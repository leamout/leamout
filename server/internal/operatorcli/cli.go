package operatorcli

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

type BuildInfo struct {
	Version string
	Commit  string
	BuiltAt string
}

type Config struct {
	RepoRoot    string
	EnvFile     string
	ComposeFile string
	CertDir     string
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, build BuildInfo) int {
	cfg := Config{
		RepoRoot:    envOr("LEAMOUT_ROOT", "."),
		EnvFile:     envOr("LEAMOUT_ENV_FILE", ".env"),
		ComposeFile: envOr("LEAMOUT_COMPOSE_FILE", "deploy/compose.yaml"),
		CertDir:     envOr("LEAMOUT_CERT_DIR", "deploy/certs"),
	}

	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "--help", "-h":
		printHelp(stdout)
		return 0
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "leamout %s\ncommit: %s\nbuilt: %s\n", build.Version, build.Commit, build.BuiltAt)
		return 0
	case "init":
		return runInit(stdout, stderr, cfg)
	case "up":
		return runScript(ctx, stdout, stderr, cfg, "scripts/deploy/up.sh")
	case "down":
		return runCompose(ctx, stdout, stderr, cfg, "down")
	case "status":
		return runCompose(ctx, stdout, stderr, cfg, "ps")
	case "logs":
		composeArgs := []string{"logs", "-f", "--tail=200"}
		composeArgs = append(composeArgs, args[1:]...)
		return runCompose(ctx, stdout, stderr, cfg, composeArgs...)
	case "doctor":
		return runDoctor(ctx, stdout, stderr, cfg)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return 2
	}
}

func runInit(stdout, stderr io.Writer, cfg Config) int {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		fmt.Fprintf(stderr, "unsupported host: %s/%s; Phase 2 supports linux/amd64\n", runtime.GOOS, runtime.GOARCH)
		return 1
	}

	root, err := filepath.Abs(cfg.RepoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "resolve Leamout root: %v\n", err)
		return 1
	}
	if _, err := os.Stat(filepath.Join(root, cfg.ComposeFile)); err != nil {
		fmt.Fprintf(stderr, "Leamout deployment assets not found under %s: %v\n", root, err)
		return 1
	}

	fmt.Fprintf(stdout, "Leamout CLI initialized for deployment assets at %s\n", root)
	fmt.Fprintln(stdout, "Phase 2 does not yet generate production secrets, TLS, or commercial activation state.")
	fmt.Fprintln(stdout, "Next: configure the existing deployment environment, then run `leamout doctor` and `leamout up`.")
	return 0
}

func runDoctor(ctx context.Context, stdout, stderr io.Writer, cfg Config) int {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		fmt.Fprintf(stderr, "unsupported host: %s/%s; Phase 2 supports linux/amd64\n", runtime.GOOS, runtime.GOARCH)
		return 1
	}

	for _, name := range []string{"docker", "sh", "curl"} {
		if _, err := exec.LookPath(name); err != nil {
			fmt.Fprintf(stderr, "missing prerequisite: %s\n", name)
			return 1
		}
	}
	fmt.Fprintln(stdout, "✓ linux/amd64 host")
	fmt.Fprintln(stdout, "✓ docker, sh, and curl available")

	if code := runScript(ctx, stdout, stderr, cfg, "scripts/deploy/preflight.sh"); code != 0 {
		return code
	}
	fmt.Fprintln(stdout, "Leamout doctor passed.")
	return 0
}

func runScript(ctx context.Context, stdout, stderr io.Writer, cfg Config, rel string) int {
	root, err := filepath.Abs(cfg.RepoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "resolve Leamout root: %v\n", err)
		return 1
	}
	path := filepath.Join(root, rel)
	if _, err := os.Stat(path); err != nil {
		fmt.Fprintf(stderr, "required deployment primitive missing: %s\n", path)
		return 1
	}

	cmd := exec.CommandContext(ctx, "sh", path)
	cmd.Dir = root
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(),
		"ENV_FILE="+cfg.EnvFile,
		"COMPOSE_FILE="+cfg.ComposeFile,
		"CERT_DIR="+cfg.CertDir,
	)
	return exitCode(cmd.Run(), stderr)
}

func runCompose(ctx context.Context, stdout, stderr io.Writer, cfg Config, args ...string) int {
	root, err := filepath.Abs(cfg.RepoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "resolve Leamout root: %v\n", err)
		return 1
	}
	base := []string{"compose", "--env-file", cfg.EnvFile, "-f", cfg.ComposeFile}
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
		fmt.Fprintf(stderr, "%v\n", err)
	}
	return 1
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `Leamout self-hosted operator CLI

Usage:
  leamout <command>

Commands:
  init       Validate the Phase 2 installation foundation
  up         Start the existing Leamout deployment stack
  down       Stop the existing Leamout deployment stack
  status     Show deployment service status
  logs       Follow deployment logs
  doctor     Validate host prerequisites and deployment preflight
  version    Print CLI build information
  help       Show this help

Environment overrides:
  LEAMOUT_ROOT          deployment asset root (default: .)
  LEAMOUT_ENV_FILE      environment file (default: .env)
  LEAMOUT_COMPOSE_FILE  compose file (default: deploy/compose.yaml)
  LEAMOUT_CERT_DIR      certificate directory (default: deploy/certs)`)
}
