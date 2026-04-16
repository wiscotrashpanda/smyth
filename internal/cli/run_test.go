package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	if err := Run([]string{"help"}, &stdout); err != nil {
		t.Fatalf("Run(help) returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Hello from Smyth.") {
		t.Fatalf("expected help output to contain greeting, got %q", output)
	}
}

func TestRunNoArgsShowsHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	if err := Run(nil, &stdout); err != nil {
		t.Fatalf("Run(nil) returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output to contain usage, got %q", output)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	err := Run([]string{"nope"}, &stdout)
	if err == nil {
		t.Fatal("expected unknown command error")
	}

	if !strings.Contains(err.Error(), "unknown command: nope") {
		t.Fatalf("expected unknown command error, got %v", err)
	}
}
