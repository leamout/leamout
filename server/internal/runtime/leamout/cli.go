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
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version string
	Commit  string
	BuiltAt string
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, build BuildInfo) int {
	root := newRootCommand(ctx, stdout, stderr, build)
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
		var status commandStatus
		if errors.As(err, &status) {
			return status.code
		}
		if len(args) > 0 && strings.HasPrefix(err.Error(), "unknown command") {
			writef(stderr, "unknown command: %s\n\n", args[0])
		} else {
			writef(stderr, "%v\n\n", err)
		}
		root.SetOut(stderr)
		_ = root.Help()
		return 2
	}
	return 0
}

type commandStatus struct{ code int }

func (e commandStatus) Error() string { return fmt.Sprintf("command exited with status %d", e.code) }

func newRootCommand(ctx context.Context, stdout, stderr io.Writer, build BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:           "leamout",
		Short:         "Operate a sovereign Leamout Self-Hosted deployment",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if version, _ := cmd.Flags().GetBool("version"); version {
				printVersion(stdout, build)
				return nil
			}
			return cmd.Help()
		},
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.CompletionOptions.DisableDefaultCmd = true
	root.SuggestionsMinimumDistance = 1
	root.Flags().BoolP("version", "v", false, "print CLI build information")

	addCommand := func(use, summary string, run func([]string) int) {
		root.AddCommand(&cobra.Command{
			Use:                use,
			Short:              summary,
			DisableFlagParsing: true,
			RunE: func(_ *cobra.Command, args []string) error {
				if code := run(args); code != 0 {
					return commandStatus{code: code}
				}
				return nil
			},
		})
	}

	root.AddCommand(&cobra.Command{
		Use:     "version",
		Aliases: []string{"ver"},
		Short:   "Print CLI build information",
		Args:    cobra.NoArgs,
		Run:     func(*cobra.Command, []string) { printVersion(stdout, build) },
	})
	addCommand("init", "Initialize the local deployment", func([]string) int {
		return runWithSpinner(stderr, "Initializing deployment", func() int { return runInit(stdout, stderr, build.Version) })
	})
	addCommand("up", "Start the installed runtime", func(args []string) int {
		if len(args) != 0 {
			writeln(stderr, "usage: leamout up")
			return 2
		}
		return runLicensedInstalledCompose(ctx, stdout, stderr, "up", "-d")
	})
	addCommand("down", "Stop the installed runtime", func(args []string) int {
		if len(args) != 0 {
			writeln(stderr, "usage: leamout down")
			return 2
		}
		return runInstalledCompose(ctx, stdout, stderr, "down")
	})
	addCommand("status", "Show runtime service status", func(args []string) int {
		if len(args) != 0 {
			writeln(stderr, "usage: leamout status")
			return 2
		}
		return runInstalledCompose(ctx, stdout, stderr, "ps")
	})
	addCommand("logs [service...]", "Follow runtime logs", func(args []string) int {
		return runInstalledCompose(ctx, stdout, stderr, append([]string{"logs", "-f", "--tail=200"}, args...)...)
	})
	addCommand("doctor", "Validate the local deployment", func(args []string) int {
		if len(args) != 0 {
			writeln(stderr, "usage: leamout doctor")
			return 2
		}
		return runWithSpinner(stderr, "Checking deployment", func() int { return runDoctor(ctx, stdout, stderr) })
	})
	addCommand("license", "Install or verify an offline license", func(args []string) int { return runLicense(stdout, stderr, args) })
	addCommand("certs", "Install or verify TLS certificates", func(args []string) int { return runCerts(stdout, stderr, args) })
	addCommand("backup", "Create a portable deployment backup", func(args []string) int { return runBackup(ctx, stdout, stderr, args) })
	addCommand("restore", "Restore a deployment backup", func(args []string) int {
		if len(args) == 1 {
			if !confirmRestore(args[0], stdout, stderr) {
				writeln(stderr, "restore cancelled")
				return 1
			}
			args = []string{"--force", args[0]}
		}
		return runRestore(ctx, stdout, stderr, args)
	})
	addCommand("update", "Install or roll back a runtime update", func(args []string) int { return runUpdate(ctx, stdout, stderr, args, build.Version) })
	return root
}

func printVersion(w io.Writer, build BuildInfo) {
	label := color.New(color.FgCyan, color.Bold)
	_, _ = label.Fprintf(w, "leamout %s\n", build.Version)
	writef(w, "commit: %s\nbuilt: %s\n", build.Commit, build.BuiltAt)
}

func runWithSpinner(w io.Writer, message string, run func() int) int {
	file, interactive := w.(*os.File)
	if !interactive {
		return run()
	}
	info, err := file.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return run()
	}
	progress := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(w))
	progress.Suffix = " " + message
	progress.Start()
	code := run()
	progress.Stop()
	return code
}

func confirmRestore(path string, stdout, stderr io.Writer) bool {
	out, outOK := stdout.(*os.File)
	if !outOK {
		return false
	}
	confirmed := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Restore %s? Existing deployment state will be replaced.", path),
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirmed, survey.WithStdio(os.Stdin, out, stderr)); err != nil {
		writef(stderr, "confirm restore: %v\n", err)
		return false
	}
	return confirmed
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

	state, err := ensureDeploymentIdentity("/var/lib/leamout/deployment.json")
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
	if _, err := validateInstalledLicense("/var/lib/leamout/deployment.json", "/etc/leamout/license", time.Now().UTC()); err != nil {
		writef(stderr, "self-hosted license: %v\n", err)
		return 1
	}
	if err := validateCertificates("/etc/leamout/certs", "", time.Now().UTC()); err != nil {
		writef(stderr, "self-hosted TLS: %v\n", err)
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
	writeln(stdout, "✓ Self-Hosted license valid")
	writeln(stdout, "✓ TLS certificates valid")
	writeln(stdout, "✓ Production runtime installed")
	writeln(stdout, "✓ Runtime configuration valid")
	writeln(stdout, "Leamout doctor passed.")
	return 0
}

func runLicensedInstalledCompose(ctx context.Context, stdout, stderr io.Writer, args ...string) int {
	if _, err := validateInstalledLicense("/var/lib/leamout/deployment.json", "/etc/leamout/license", time.Now().UTC()); err != nil {
		writef(stderr, "self-hosted license: %v\nInstall a valid license before starting Leamout.\n", err)
		return 1
	}
	if err := validateCertificates("/etc/leamout/certs", "", time.Now().UTC()); err != nil {
		writef(stderr, "self-hosted TLS: %v\nInstall valid certificates before starting Leamout.\n", err)
		return 1
	}
	return runInstalledCompose(ctx, stdout, stderr, args...)
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
