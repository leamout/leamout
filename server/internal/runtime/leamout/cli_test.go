package operatorcli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"version"}, &stdout, &stderr, BuildInfo{
		Version: "1.2.3",
		Commit:  "abc123",
		BuiltAt: "2026-09-03T00:00:00Z",
	})
	if code != 0 {
		t.Fatalf("Run returned %d: %s", code, stderr.String())
	}
	for _, want := range []string{"leamout 1.2.3", "commit: abc123", "built: 2026-09-03T00:00:00Z"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("version output missing %q: %s", want, stdout.String())
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"nope"}, &stdout, &stderr, BuildInfo{})
	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown command: nope") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}
