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

const (
	selfHostedBuildServices  = "server worker opensips freeswitch rtpengine"
	selfHostedDeployServices = "server worker opensips coturn freeswitch rtpengine postgres redis nats"
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
	ConfigDir   string
	StateDir    string
	LogDir      string
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, build BuildInfo) int {
	cfg := Config{
		RepoRoot:    envOr("LEAMOUT_ROOT", "."),
		EnvFile:     envOr("LEAMOUT_ENV_FILE", ".env"),
		ComposeFile: envOr("LEAMOUT_COMPOSE_FILE", "deploy/compose.yaml"),
		CertDir:     envOr("LEAMOUT_CERT_DIR", "deploy/certs"),
		ConfigDir:   envOr("LEAMOUT_CONFIG_DIR", "/etc/leamout"),
		StateDir:    envOr("LEAMOUT_STATE_DIR", "/var/lib/leamout"),
		LogDir:      envOr("LEAMOUT_LOG_DIR", "/var/log/leamout"),
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
		writef(stdout, "leamout %s\ncommit: %s\nbuilt: %s\n", build.Version, build.Commit, build.BuiltAt)
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
		writef(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return 2
	}
}

func runInit(stdout, stderr io.Writer, cfg Config) int {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		writef(stderr, "unsupported host: %s/%s; Phase 2 supports linux/amd64\n", runtime.GOOS, runtime.GOARCH)
		return 1
	}

	for _, dir := range []string{cfg.ConfigDir, cfg.StateDir, cfg.LogDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			writef(stderr, "create %s: %v\n", dir, err)
			return 1
		}
	}

	writeln(stdout, "✓ Leamout base directories initialized")
	writeln(stdout, "Phase 2 does not yet generate production secrets, TLS, deployment identity, or commercial activation state.")

	root, err := filepath.Abs(cfg.RepoRoot)
	if err == nil {
		if _, statErr := os.Stat(filepath.Join(root, cfg.ComposeFile)); statErr == nil {
			writef(stdout, "✓ Existing deployment assets detected at %s\n", root)
			writeln(stdout, "Next: configure the deployment environment, then run `leamout doctor` and `leamout up`.")
			return 0
		}
	}

	writeln(stdout, "Runtime deployment assets are not installed yet; Phase 3 will make clean-host runtime initialization self-contained.")
	return 0
}

func runDoctor(ctx context.Context, stdout, stderr io.Writer, cfg Config) int {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		writef(stderr, "unsupported host: %s/%s; Phase 2 supports linux/amd64\n", runtime.GOOS, runtime.GOARCH)
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

	if code := runScript(ctx, stdout, stderr, cfg, "scripts/deploy/preflight.sh"); code != 0 {
		return code
	}
	writeln(stdout, "Leamout doctor passed.")
	return 0
}

func runScript(ctx context.Context, stdout, stderr io.Writer, cfg Config, rel string) int {
	root, err := filepath.Abs(cfg.RepoRoot)
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
		"ENV_FILE="+cfg.EnvFile,
		"COMPOSE_FILE="+cfg.ComposeFile,
		"CERT_DIR="+cfg.CertDir,
		"BUILD_SERVICES="+selfHostedBuildServices,
		"DEPLOY_SERVICES="+selfHostedDeployServices,
	)
	return exitCode(cmd.Run(), stderr)
}

func runCompose(ctx context.Context, stdout, stderr io.Writer, cfg Config, args ...string) int {
	root, err := filepath.Abs(cfg.RepoRoot)
	if err != nil {
		writef(stderr, "resolve Leamout root: %v\n", err)
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
		writef(stderr, "%v\n", err)
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
	writeln(w, `Leamout self-hosted operator CLI

Usage:
  leamout <command>

Commands:
  init       Initialize the Phase 2 local CLI foundation
  up         Start self-hosted runtime services through existing deployment primitives
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
  LEAMOUT_CERT_DIR      certificate directory (default: deploy/certs)
  LEAMOUT_CONFIG_DIR    local configuration directory (default: /etc/leamout)
  LEAMOUT_STATE_DIR     durable state directory (default: /var/lib/leamout)
  LEAMOUT_LOG_DIR       local log directory (default: /var/log/leamout)`)
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
