package leamout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func drainInstalledRuntime(ctx context.Context, stdout, stderr io.Writer, timeout, poll time.Duration) error {
	if err := runDrainControl(ctx, "opensips", "/usr/local/bin/leamout-opensips-drain", "enable"); err != nil {
		return err
	}
	if err := runDrainControl(ctx, "freeswitch", "/usr/local/bin/leamout-freeswitch-drain", "enable"); err != nil {
		_ = runDrainControl(ctx, "opensips", "/usr/local/bin/leamout-opensips-drain", "resume")
		return err
	}

	deadline := time.Now().Add(timeout)
	for {
		dialogs, err := drainCount(ctx, "opensips", "/usr/local/bin/leamout-opensips-drain", "dialogs")
		if err != nil {
			_ = resumeInstalledRuntime(ctx, stderr)
			return err
		}
		channels, err := drainCount(ctx, "freeswitch", "/usr/local/bin/leamout-freeswitch-drain", "channels")
		if err != nil {
			_ = resumeInstalledRuntime(ctx, stderr)
			return err
		}
		writef(stdout, "Drain progress: OpenSIPS dialogs=%d FreeSWITCH channels=%d\n", dialogs, channels)
		if dialogs == 0 && channels == 0 {
			return nil
		}
		if time.Now().Add(poll).After(deadline) {
			_ = resumeInstalledRuntime(ctx, stderr)
			return errors.New("deadline reached with active telecom sessions")
		}
		select {
		case <-ctx.Done():
			_ = resumeInstalledRuntime(context.Background(), stderr)
			return ctx.Err()
		case <-time.After(poll):
		}
	}
}

func runDrainControl(ctx context.Context, service, command, action string) error {
	cmd := installedComposeCommand(ctx, "exec", "-T", service, command, action)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %s: %w: %s", service, action, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func drainCount(ctx context.Context, service, command, action string) (int, error) {
	cmd := installedComposeCommand(ctx, "exec", "-T", service, command, action)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("query %s %s: %w: %s", service, action, err, strings.TrimSpace(string(output)))
	}
	count, err := parseDrainCount(output)
	if err != nil || count < 0 {
		return 0, fmt.Errorf("query %s %s returned invalid count %q", service, action, strings.TrimSpace(string(output)))
	}
	return count, nil
}

func parseDrainCount(output []byte) (int, error) {
	return strconv.Atoi(strings.TrimSpace(string(output)))
}

func resumeInstalledRuntime(ctx context.Context, stderr io.Writer) error {
	var failures []error
	for _, control := range []struct{ service, command string }{
		{"freeswitch", "/usr/local/bin/leamout-freeswitch-drain"},
		{"opensips", "/usr/local/bin/leamout-opensips-drain"},
	} {
		if err := runDrainControl(ctx, control.service, control.command, "resume"); err != nil {
			writef(stderr, "resume %s admission: %v\n", control.service, err)
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
