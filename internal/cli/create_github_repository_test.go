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
		"", // visibility -> unmanaged
		"", // description -> unmanaged
		"", // homepage -> unmanaged
		"", // default branch -> unmanaged
		"", // auto-init -> false
		"", // topics -> unmanaged
		"", // features -> skipped
		"", // merge policy -> skipped
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

	if manifest.Spec.Visibility != nil {
		t.Errorf("spec.visibility: got %v, want nil", manifest.Spec.Visibility)
	}

	if manifest.Spec.Description != nil {
		t.Errorf("spec.description: got %v, want nil", manifest.Spec.Description)
	}

	if manifest.Spec.Homepage != nil {
		t.Errorf("spec.homepage: got %v, want nil", manifest.Spec.Homepage)
	}

	if manifest.Spec.DefaultBranch != nil {
		t.Errorf("spec.defaultBranch: got %v, want nil", manifest.Spec.DefaultBranch)
	}

	if manifest.Spec.AutoInit {
		t.Errorf("spec.autoInit: got true, want false")
	}

	if manifest.Spec.Topics != nil {
		t.Errorf("spec.topics: got %v, want nil", manifest.Spec.Topics)
	}

	if manifest.Spec.Features != nil {
		t.Errorf("spec.features: got %#v, want nil", manifest.Spec.Features)
	}

	if manifest.Spec.MergePolicy != nil {
		t.Errorf("spec.mergePolicy: got %#v, want nil", manifest.Spec.MergePolicy)
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
		"y",
		"y",
		"n",
		"n",
		"y",
		"y",
		"n",
		"y",
		"y",
		"n",
		"y",
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

	assertStringPointerValue(t, "spec.visibility", manifest.Spec.Visibility, "public")
	assertStringPointerValue(t, "spec.description", manifest.Spec.Description, "An example description")
	assertStringPointerValue(t, "spec.homepage", manifest.Spec.Homepage, "https://example.com")
	assertStringPointerValue(t, "spec.defaultBranch", manifest.Spec.DefaultBranch, "trunk")

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

	if manifest.Spec.Features == nil {
		t.Fatal("spec.features: got nil, want populated feature toggles")
	}

	assertBoolPointerValue(t, "spec.features.hasIssues", manifest.Spec.Features.HasIssues, true)
	assertBoolPointerValue(t, "spec.features.hasProjects", manifest.Spec.Features.HasProjects, false)
	assertBoolPointerValue(t, "spec.features.hasWiki", manifest.Spec.Features.HasWiki, false)

	if manifest.Spec.MergePolicy == nil {
		t.Fatal("spec.mergePolicy: got nil, want populated merge policy")
	}

	assertBoolPointerValue(t, "spec.mergePolicy.allowSquashMerge", manifest.Spec.MergePolicy.AllowSquashMerge, true)
	assertBoolPointerValue(t, "spec.mergePolicy.allowMergeCommit", manifest.Spec.MergePolicy.AllowMergeCommit, false)
	assertBoolPointerValue(t, "spec.mergePolicy.allowRebaseMerge", manifest.Spec.MergePolicy.AllowRebaseMerge, true)
	assertBoolPointerValue(t, "spec.mergePolicy.allowAutoMerge", manifest.Spec.MergePolicy.AllowAutoMerge, true)
	assertBoolPointerValue(t, "spec.mergePolicy.allowUpdateBranch", manifest.Spec.MergePolicy.AllowUpdateBranch, false)
	assertBoolPointerValue(t, "spec.mergePolicy.deleteBranchOnMerge", manifest.Spec.MergePolicy.DeleteBranchOnMerge, true)
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
		"", "", "", "", "", "", "", "", "",
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
		"", "", "", "", "", "", "", "",
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
		"", "", "", "", "", "", "", "",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "github-repo", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	if !strings.Contains(stdout.String(), "must be one of") {
		t.Fatalf("expected visibility validation message, got:\n%s", stdout.String())
	}
}

func assertStringPointerValue(t *testing.T, field string, got *string, want string) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s: got nil, want %q", field, want)
	}

	if *got != want {
		t.Fatalf("%s: got %q, want %q", field, *got, want)
	}
}

func assertBoolPointerValue(t *testing.T, field string, got *bool, want bool) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s: got nil, want %t", field, want)
	}

	if *got != want {
		t.Fatalf("%s: got %t, want %t", field, *got, want)
	}
}
