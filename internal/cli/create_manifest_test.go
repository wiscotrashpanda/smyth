package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCreateManifestNoTypeShowsHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest"}, nil, &stdout); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Manifest types:") {
		t.Fatalf("expected help output to list manifest types, got:\n%s", output)
	}

	if !strings.Contains(output, "github-repo") {
		t.Fatalf("expected help output to mention github-repo, got:\n%s", output)
	}

	if !strings.Contains(output, "hcp-terraform-workspace") {
		t.Fatalf("expected help output to mention hcp-terraform-workspace, got:\n%s", output)
	}
}

func TestCreateManifestUnknownType(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	err := Run([]string{"create-manifest", "s3-bucket"}, nil, &stdout)
	if err == nil {
		t.Fatal("expected error for unknown manifest type")
	}

	if !strings.Contains(err.Error(), "unknown manifest type: s3-bucket") {
		t.Fatalf("unexpected error: %v", err)
	}
}
