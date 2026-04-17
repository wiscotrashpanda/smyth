package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1alpha1 "github.com/emkaytec/alloy/manifest/v1alpha1"
	"gopkg.in/yaml.v3"
)

func TestCreateHCPTerraformWorkspaceUsesDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	input := strings.Join([]string{
		"example-org",
		"example-workspace",
		"", // projectID
		"", // description
		"", // terraformVersion
		"", // workingDirectory
		"", // executionMode
		"", // autoApply
		"", // queueAllRuns
		"", // fileTriggersEnabled
		"", // speculativeEnabled
		"", // tags
		"", // triggerPatterns
		"", // triggerPrefixes
		"", // remoteStateConsumerIDs
		"", // variableSetIDs
		"", // vcsRepo
		"", // variables
		"",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "hcp-terraform-workspace", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	manifestPath := filepath.Join(dir, "example-org-example-workspace.manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read generated manifest: %v", err)
	}

	var manifest v1alpha1.HCPTerraformWorkspaceManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v\ncontents: %s", err, string(data))
	}

	if manifest.APIVersion != v1alpha1.APIVersion {
		t.Errorf("apiVersion: got %q, want %q", manifest.APIVersion, v1alpha1.APIVersion)
	}

	if manifest.Kind != v1alpha1.KindHCPTerraformWorkspace {
		t.Errorf("kind: got %q, want %q", manifest.Kind, v1alpha1.KindHCPTerraformWorkspace)
	}

	if manifest.Metadata.Name != "example-org-example-workspace" {
		t.Errorf("metadata.name: got %q, want example-org-example-workspace", manifest.Metadata.Name)
	}

	if manifest.Spec.Organization != "example-org" {
		t.Errorf("spec.organization: got %q, want example-org", manifest.Spec.Organization)
	}

	if manifest.Spec.Name != "example-workspace" {
		t.Errorf("spec.name: got %q, want example-workspace", manifest.Spec.Name)
	}

	if manifest.Spec.ProjectID != nil {
		t.Errorf("spec.projectID: got %v, want nil", manifest.Spec.ProjectID)
	}

	if manifest.Spec.ExecutionMode != nil {
		t.Errorf("spec.executionMode: got %v, want nil", manifest.Spec.ExecutionMode)
	}

	if manifest.Spec.Tags != nil {
		t.Errorf("spec.tags: got %v, want nil", manifest.Spec.Tags)
	}

	if manifest.Spec.VCSRepo != nil {
		t.Errorf("spec.vcsRepo: got %#v, want nil", manifest.Spec.VCSRepo)
	}

	if manifest.Spec.Variables != nil {
		t.Errorf("spec.variables: got %#v, want nil", manifest.Spec.Variables)
	}

	if err := manifest.Validate(); err != nil {
		t.Errorf("generated manifest failed validation: %v", err)
	}
}

func TestCreateHCPTerraformWorkspaceCollectsOptionalFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	input := strings.Join([]string{
		"example-org",
		"example-workspace",
		"prj-123456",
		"Managed workspace",
		"1.14.8",
		"terraform",
		"agent",
		"apool-123456",
		"y", // autoApply
		"n", // queueAllRuns
		"y", // fileTriggersEnabled
		"y", // speculativeEnabled
		"platform, hcp, platform",
		"terraform/**/*.tf, modules/**/*.tf",
		"terraform/, modules/",
		"ws-1, ws-2",
		"varset-1, varset-2",
		"y", // configure vcs
		"example-org/example-repo",
		"ot-123456",
		"main",
		"n",     // ingressSubmodules
		"^v.*$", // tagsRegex
		"y",     // add variables
		"AWS_REGION",
		"env",
		"us-east-1",
		"Default region",
		"",  // sensitive omitted
		"y", // add another
		"account_id",
		"terraform",
		"\"123456789012\"",
		"",
		"n", // sensitive false
		"y", // hcl true
		"n", // stop adding
		"",
	}, "\n")

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "hcp-terraform-workspace", "--dir", dir}, strings.NewReader(input), &stdout); err != nil {
		t.Fatalf("Run returned error: %v\noutput: %s", err, stdout.String())
	}

	manifestPath := filepath.Join(dir, "example-org-example-workspace.manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read generated manifest: %v", err)
	}

	var manifest v1alpha1.HCPTerraformWorkspaceManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v\ncontents: %s", err, string(data))
	}

	assertStringPointerValue(t, "spec.projectID", manifest.Spec.ProjectID, "prj-123456")
	assertStringPointerValue(t, "spec.description", manifest.Spec.Description, "Managed workspace")
	assertStringPointerValue(t, "spec.terraformVersion", manifest.Spec.TerraformVersion, "1.14.8")
	assertStringPointerValue(t, "spec.workingDirectory", manifest.Spec.WorkingDirectory, "terraform")
	assertStringPointerValue(t, "spec.executionMode", manifest.Spec.ExecutionMode, "agent")
	assertStringPointerValue(t, "spec.agentPoolID", manifest.Spec.AgentPoolID, "apool-123456")

	assertBoolPointerValue(t, "spec.autoApply", manifest.Spec.AutoApply, true)
	assertBoolPointerValue(t, "spec.queueAllRuns", manifest.Spec.QueueAllRuns, false)
	assertBoolPointerValue(t, "spec.fileTriggersEnabled", manifest.Spec.FileTriggersEnabled, true)
	assertBoolPointerValue(t, "spec.speculativeEnabled", manifest.Spec.SpeculativeEnabled, true)

	wantTags := []string{"platform", "hcp"}
	if len(manifest.Spec.Tags) != len(wantTags) {
		t.Fatalf("spec.tags: got %v, want %v", manifest.Spec.Tags, wantTags)
	}
	for i, tag := range wantTags {
		if manifest.Spec.Tags[i] != tag {
			t.Errorf("spec.tags[%d]: got %q, want %q", i, manifest.Spec.Tags[i], tag)
		}
	}

	wantPatterns := []string{"terraform/**/*.tf", "modules/**/*.tf"}
	if len(manifest.Spec.TriggerPatterns) != len(wantPatterns) {
		t.Fatalf("spec.triggerPatterns: got %v, want %v", manifest.Spec.TriggerPatterns, wantPatterns)
	}

	wantPrefixes := []string{"terraform/", "modules/"}
	if len(manifest.Spec.TriggerPrefixes) != len(wantPrefixes) {
		t.Fatalf("spec.triggerPrefixes: got %v, want %v", manifest.Spec.TriggerPrefixes, wantPrefixes)
	}

	wantConsumers := []string{"ws-1", "ws-2"}
	if len(manifest.Spec.RemoteStateConsumerIDs) != len(wantConsumers) {
		t.Fatalf("spec.remoteStateConsumerIDs: got %v, want %v", manifest.Spec.RemoteStateConsumerIDs, wantConsumers)
	}

	wantVarSets := []string{"varset-1", "varset-2"}
	if len(manifest.Spec.VariableSetIDs) != len(wantVarSets) {
		t.Fatalf("spec.variableSetIDs: got %v, want %v", manifest.Spec.VariableSetIDs, wantVarSets)
	}

	if manifest.Spec.VCSRepo == nil {
		t.Fatal("spec.vcsRepo: got nil, want populated VCS repo")
	}
	assertStringPointerValue(t, "spec.vcsRepo.identifier", manifest.Spec.VCSRepo.Identifier, "example-org/example-repo")
	assertStringPointerValue(t, "spec.vcsRepo.oauthTokenID", manifest.Spec.VCSRepo.OAuthTokenID, "ot-123456")
	assertStringPointerValue(t, "spec.vcsRepo.branch", manifest.Spec.VCSRepo.Branch, "main")
	assertBoolPointerValue(t, "spec.vcsRepo.ingressSubmodules", manifest.Spec.VCSRepo.IngressSubmodules, false)
	assertStringPointerValue(t, "spec.vcsRepo.tagsRegex", manifest.Spec.VCSRepo.TagsRegex, "^v.*$")

	if len(manifest.Spec.Variables) != 2 {
		t.Fatalf("spec.variables: got %d, want 2", len(manifest.Spec.Variables))
	}

	if manifest.Spec.Variables[0].Key != "AWS_REGION" || manifest.Spec.Variables[0].Category != "env" {
		t.Fatalf("spec.variables[0]: got %#v", manifest.Spec.Variables[0])
	}
	if manifest.Spec.Variables[0].Description == nil || *manifest.Spec.Variables[0].Description != "Default region" {
		t.Fatalf("spec.variables[0].description: got %#v", manifest.Spec.Variables[0].Description)
	}
	if manifest.Spec.Variables[1].Key != "account_id" || manifest.Spec.Variables[1].Category != "terraform" {
		t.Fatalf("spec.variables[1]: got %#v", manifest.Spec.Variables[1])
	}
	assertBoolPointerValue(t, "spec.variables[1].sensitive", manifest.Spec.Variables[1].Sensitive, false)
	assertBoolPointerValue(t, "spec.variables[1].hcl", manifest.Spec.Variables[1].HCL, true)

	if err := manifest.Validate(); err != nil {
		t.Errorf("generated manifest failed validation: %v", err)
	}
}

func TestCreateHCPTerraformWorkspaceHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	if err := Run([]string{"create-manifest", "hcp-terraform-workspace", "--help"}, nil, &stdout); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "smyth create-manifest hcp-terraform-workspace") {
		t.Fatalf("expected help output to mention command usage, got:\n%s", output)
	}
}
