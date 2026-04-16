package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	v1alpha1 "github.com/emkaytec/alloy/manifest/v1alpha1"
	"gopkg.in/yaml.v3"
)

func TestCreateGitHubRepositoryUsesDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Owner, repo name, then blank lines to accept every default.
	input := strings.Join([]string{
		"example-org",
		"example-repo",
		"", // visibility -> "private"
		"", // description -> ""
		"", // homepage -> ""
		"", // default branch -> "main"
		"", // auto-init -> false
		"", // topics -> none
		"",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "github-repo", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	manifestPath := filepath.Join(dir, "example-org-example-repo.manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read generated manifest: %v", err)
	}

	var manifest v1alpha1.GitHubRepositoryManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v\ncontents: %s", err, string(data))
	}

	if manifest.APIVersion != v1alpha1.APIVersion {
		t.Errorf("apiVersion: got %q, want %q", manifest.APIVersion, v1alpha1.APIVersion)
	}

	if manifest.Kind != v1alpha1.KindGitHubRepository {
		t.Errorf("kind: got %q, want %q", manifest.Kind, v1alpha1.KindGitHubRepository)
	}

	if manifest.Metadata.Name != "example-org-example-repo" {
		t.Errorf("metadata.name: got %q, want example-org-example-repo", manifest.Metadata.Name)
	}

	if manifest.Spec.Owner != "example-org" {
		t.Errorf("spec.owner: got %q, want example-org", manifest.Spec.Owner)
	}

	if manifest.Spec.Name != "example-repo" {
		t.Errorf("spec.name: got %q, want example-repo", manifest.Spec.Name)
	}

	if got := derefString(manifest.Spec.Visibility); got != "private" {
		t.Errorf("spec.visibility: got %q, want private", got)
	}

	if got := derefString(manifest.Spec.DefaultBranch); got != "main" {
		t.Errorf("spec.defaultBranch: got %q, want main", got)
	}

	if manifest.Spec.Description != nil {
		t.Errorf("spec.description: got %q, want nil", *manifest.Spec.Description)
	}

	if manifest.Spec.Homepage != nil {
		t.Errorf("spec.homepage: got %q, want nil", *manifest.Spec.Homepage)
	}

	if manifest.Spec.AutoInit {
		t.Errorf("spec.autoInit: got true, want false")
	}

	if len(manifest.Spec.Topics) != 0 {
		t.Errorf("spec.topics: got %v, want empty", manifest.Spec.Topics)
	}

	if err := manifest.Validate(); err != nil {
		t.Errorf("generated manifest failed validation: %v", err)
	}
}

func TestCreateGitHubRepositoryCollectsOptionalFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	input := strings.Join([]string{
		"example-org",
		"example-repo",
		"public",
		"An example description",
		"https://example.com",
		"trunk",
		"y",
		"platform, observability, platform",
		"",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "github-repo", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	manifestPath := filepath.Join(dir, "example-org-example-repo.manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read generated manifest: %v", err)
	}

	var manifest v1alpha1.GitHubRepositoryManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v\ncontents: %s", err, string(data))
	}

	if manifest.Metadata.Name != "example-org-example-repo" {
		t.Errorf("metadata.name: got %q, want example-org-example-repo", manifest.Metadata.Name)
	}

	if got := derefString(manifest.Spec.Visibility); got != "public" {
		t.Errorf("spec.visibility: got %q, want public", got)
	}

	if got := derefString(manifest.Spec.Description); got != "An example description" {
		t.Errorf("spec.description: got %q", got)
	}

	if got := derefString(manifest.Spec.Homepage); got != "https://example.com" {
		t.Errorf("spec.homepage: got %q", got)
	}

	if got := derefString(manifest.Spec.DefaultBranch); got != "trunk" {
		t.Errorf("spec.defaultBranch: got %q, want trunk", got)
	}

	if !manifest.Spec.AutoInit {
		t.Errorf("spec.autoInit: got false, want true")
	}

	wantTopics := []string{"platform", "observability"}
	if len(manifest.Spec.Topics) != len(wantTopics) {
		t.Fatalf("spec.topics: got %v, want %v", manifest.Spec.Topics, wantTopics)
	}

	for i, topic := range wantTopics {
		if manifest.Spec.Topics[i] != topic {
			t.Errorf("spec.topics[%d]: got %q, want %q", i, manifest.Spec.Topics[i], topic)
		}
	}
}

func TestCreateGitHubRepositoryAbortsWhenManifestExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	existingPath := filepath.Join(dir, "example-org-example-repo.manifest.yaml")
	if err := os.WriteFile(existingPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}

	input := strings.Join([]string{
		"example-org",
		"example-repo",
		"n", // decline to continue
		"",
	}, "\n")

	var stdout bytes.Buffer

	err := Run([]string{"create-manifest", "github-repo", "--dir", dir}, strings.NewReader(input), &stdout)
	if err == nil {
		t.Fatal("expected error when user declines to continue")
	}

	if !strings.Contains(err.Error(), "aborted: manifest already exists") {
		t.Fatalf("unexpected error: %v", err)
	}

	// The pre-existing file should be left untouched.
	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("re-read existing file: %v", err)
	}

	if string(data) != "existing" {
		t.Fatalf("existing file was modified: %q", string(data))
	}

	// No stray new files should have been created in the directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	if len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}

		t.Fatalf("expected only the seeded file, got: %v", names)
	}
}

func TestCreateGitHubRepositoryAppendsSuffixWhenContinuing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	existingPath := filepath.Join(dir, "example-org-example-repo.manifest.yaml")
	if err := os.WriteFile(existingPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}

	input := strings.Join([]string{
		"example-org",
		"example-repo",
		"y", // continue despite existing manifest
		"", "", "", "", "", "", "",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "github-repo", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	suffixed := ""
	for _, entry := range entries {
		if entry.Name() == "example-org-example-repo.manifest.yaml" {
			continue
		}

		suffixed = entry.Name()
	}

	if suffixed == "" {
		t.Fatal("expected a new suffixed manifest file to be created")
	}

	pattern := regexp.MustCompile(`^example-org-example-repo-[a-z0-9]{4}\.manifest\.yaml$`)
	if !pattern.MatchString(suffixed) {
		t.Fatalf("suffixed filename %q does not match expected pattern", suffixed)
	}

	data, err := os.ReadFile(filepath.Join(dir, suffixed))
	if err != nil {
		t.Fatalf("read suffixed manifest: %v", err)
	}

	var manifest v1alpha1.GitHubRepositoryManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if manifest.Spec.Name != "example-repo" {
		t.Errorf("spec.name: got %q, want example-repo", manifest.Spec.Name)
	}

	if manifest.Metadata.Name != "example-org-example-repo" {
		t.Errorf("metadata.name: got %q, want example-org-example-repo", manifest.Metadata.Name)
	}

	// The pre-existing file should still be the original placeholder.
	seedData, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("re-read existing file: %v", err)
	}

	if string(seedData) != "existing" {
		t.Fatalf("pre-existing manifest was modified: %q", string(seedData))
	}
}

func TestCreateGitHubRepositoryNormalizesRepositoryName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	input := strings.Join([]string{
		"example-org",
		"  Repo With Spaces  ",
		"", "", "", "", "", "",
		"",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "github-repo", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	manifestPath := filepath.Join(dir, "example-org-repo-with-spaces.manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read generated manifest: %v", err)
	}

	var manifest v1alpha1.GitHubRepositoryManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v\ncontents: %s", err, string(data))
	}

	if manifest.Metadata.Name != "example-org-repo-with-spaces" {
		t.Errorf("metadata.name: got %q, want example-org-repo-with-spaces", manifest.Metadata.Name)
	}

	if manifest.Spec.Name != "repo-with-spaces" {
		t.Errorf("spec.name: got %q, want repo-with-spaces", manifest.Spec.Name)
	}

	if !strings.Contains(stdout.String(), "using repository name repo-with-spaces") {
		t.Fatalf("expected normalization message, got:\n%s", stdout.String())
	}
}

func TestCreateGitHubRepositoryValidatesVisibility(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Supply a bogus visibility on the first attempt, then accept the default.
	input := strings.Join([]string{
		"example-org",
		"example-repo",
		"hybrid",
		"private",
		"", "", "", "", "", "",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "github-repo", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	if !strings.Contains(stdout.String(), "must be one of") {
		t.Fatalf("expected visibility validation message, got:\n%s", stdout.String())
	}
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}

	return *p
}
