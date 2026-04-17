package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	v1alpha1 "github.com/emkaytec/alloy/manifest/v1alpha1"
)

// runCreateHCPTerraformWorkspace handles `smyth create-manifest hcp-terraform-workspace`.
// It prompts for the core workspace settings Smyth can author conveniently
// today, validates the manifest via alloy, and writes it to disk as YAML.
func runCreateHCPTerraformWorkspace(args []string, stdin io.Reader, stdout io.Writer) error {
	s := newStyler(stdout)

	flags := flag.NewFlagSet("create-manifest hcp-terraform-workspace", flag.ContinueOnError)
	flags.SetOutput(stdout)

	dir := flags.String("dir", ".", "Directory to write the manifest into")

	flags.Usage = func() {
		writeCreateHCPTerraformWorkspaceHelp(stdout, s)
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return err
	}

	if flags.NArg() > 0 {
		return fmt.Errorf("create-manifest hcp-terraform-workspace takes no positional arguments")
	}

	writeBanner(stdout, s)
	writeSectionHeader(stdout, s, "Authoring a HCPTerraformWorkspace manifest")
	fmt.Fprintln(stdout, s.dim("  press enter to accept the default shown in brackets"))
	fmt.Fprintln(stdout, s.dim("  fields without defaults are omitted from the manifest when left blank"))
	fmt.Fprintln(stdout)

	p := newPrompter(stdin, stdout)

	organization, err := p.askRequired("HCP Terraform organization")
	if err != nil {
		return err
	}

	name, err := askWorkspaceName(p)
	if err != nil {
		return err
	}

	filename, err := disambiguateManifestFilename(*dir, organization, name, p)
	if err != nil {
		return err
	}

	writeSectionHeader(stdout, s, "Core settings")

	projectID, err := p.askOptional("Project ID")
	if err != nil {
		return err
	}

	description, err := p.askOptional("Description")
	if err != nil {
		return err
	}

	terraformVersion, err := p.askOptional("Terraform version")
	if err != nil {
		return err
	}

	workingDirectory, err := p.askOptional("Working directory")
	if err != nil {
		return err
	}

	executionMode, err := p.askOptionalChoice("Execution mode", []string{"remote", "local", "agent"})
	if err != nil {
		return err
	}

	var agentPoolID *string
	if executionMode != nil && *executionMode == "agent" {
		agentPoolID, err = p.askRequiredAsOptional("Agent pool ID")
		if err != nil {
			return err
		}
	}

	writeSectionHeader(stdout, s, "Run behavior")
	fmt.Fprintln(stdout, s.dim("  leave toggles blank to keep them unmanaged"))

	autoApply, err := p.askOptionalBool("Auto-apply")
	if err != nil {
		return err
	}

	queueAllRuns, err := p.askOptionalBool("Queue all runs")
	if err != nil {
		return err
	}

	fileTriggersEnabled, err := p.askOptionalBool("File triggers enabled")
	if err != nil {
		return err
	}

	speculativeEnabled, err := p.askOptionalBool("Speculative runs enabled")
	if err != nil {
		return err
	}

	writeSectionHeader(stdout, s, "Workspace metadata")

	tags, err := p.askList("Tags (comma separated, optional)")
	if err != nil {
		return err
	}

	triggerPatterns, err := p.askList("Trigger patterns (comma separated, optional)")
	if err != nil {
		return err
	}

	triggerPrefixes, err := p.askList("Trigger prefixes (comma separated, optional)")
	if err != nil {
		return err
	}

	remoteStateConsumerIDs, err := p.askList("Remote state consumer workspace IDs (comma separated, optional)")
	if err != nil {
		return err
	}

	variableSetIDs, err := p.askList("Variable set IDs (comma separated, optional)")
	if err != nil {
		return err
	}

	writeSectionHeader(stdout, s, "VCS integration")
	fmt.Fprintln(stdout, s.dim("  leave this off if the workspace will not connect to a VCS repository"))

	vcsRepo, err := askHCPTerraformWorkspaceVCSRepo(p)
	if err != nil {
		return err
	}

	writeSectionHeader(stdout, s, "Workspace variables")
	fmt.Fprintln(stdout, s.dim("  add the variables Smyth should write directly into the manifest"))

	variables, err := askHCPTerraformWorkspaceVariables(p)
	if err != nil {
		return err
	}

	metadataName := defaultMetadataName(organization, name)

	manifest := v1alpha1.NewHCPTerraformWorkspaceManifest(
		v1alpha1.Metadata{Name: metadataName},
		v1alpha1.HCPTerraformWorkspaceSpec{
			Organization:           organization,
			Name:                   name,
			ProjectID:              projectID,
			Description:            description,
			TerraformVersion:       terraformVersion,
			WorkingDirectory:       workingDirectory,
			ExecutionMode:          executionMode,
			AgentPoolID:            agentPoolID,
			AutoApply:              autoApply,
			QueueAllRuns:           queueAllRuns,
			FileTriggersEnabled:    fileTriggersEnabled,
			SpeculativeEnabled:     speculativeEnabled,
			Tags:                   tags,
			TriggerPatterns:        triggerPatterns,
			TriggerPrefixes:        triggerPrefixes,
			RemoteStateConsumerIDs: remoteStateConsumerIDs,
			VariableSetIDs:         variableSetIDs,
			VCSRepo:                vcsRepo,
			Variables:              variables,
		},
	)

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("generated manifest is invalid: %w", err)
	}

	encoded, err := encodeManifest(manifest)
	if err != nil {
		return err
	}

	outputPath, err := writeManifest(*dir, filename, encoded)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout)
	writeSectionHeader(stdout, s, "Forged")
	fmt.Fprintf(stdout, "  %s wrote manifest to %s\n", s.green("✓"), s.bold(outputPath))
	fmt.Fprintf(stdout, "  %s review the file and commit it, then hand it to anvil to reconcile.\n", s.dim("›"))

	return nil
}

func writeCreateHCPTerraformWorkspaceHelp(stdout io.Writer, s *styler) {
	fmt.Fprintf(stdout, `%s
  smyth create-manifest hcp-terraform-workspace [--dir <path>]

Interactively prompts for the core fields needed to build a
HCPTerraformWorkspace manifest for Anvil's current HCP Terraform workspace
surface. Fields left blank are omitted so the generated manifest only declares
settings you want Anvil to manage. The prompt flow includes core workspace
settings, VCS repo wiring, tags, trigger settings, variable sets, remote state
consumers, and workspace variables.

%s
  %s %s
`,
		s.bold("Usage:"),
		s.bold("Flags:"),
		s.cyan("--dir <path>"),
		s.dim("Directory to write the manifest into (default: current directory)"),
	)
}

func askWorkspaceName(p *prompter) (string, error) {
	for {
		raw, err := p.askRequired("Workspace name")
		if err != nil {
			return "", err
		}

		normalized := normalizeWorkspaceName(raw)
		if normalized == "" {
			p.warn("workspace name must contain letters or numbers")
			continue
		}

		if normalized != strings.TrimSpace(raw) {
			fmt.Fprintf(
				p.writer,
				"  %s using workspace name %s\n",
				p.style.dim("›"),
				p.style.bold(normalized),
			)
		}

		return normalized, nil
	}
}

func askHCPTerraformWorkspaceVCSRepo(p *prompter) (*v1alpha1.HCPTerraformWorkspaceVCSRepoSpec, error) {
	configure, err := p.askBool("Configure VCS repository settings", false)
	if err != nil {
		return nil, err
	}

	if !configure {
		return nil, nil
	}

	identifier, err := p.askRequiredAsOptional("VCS repository identifier (owner/repo)")
	if err != nil {
		return nil, err
	}

	oauthTokenID, err := p.askRequiredAsOptional("OAuth token ID")
	if err != nil {
		return nil, err
	}

	branch, err := p.askOptional("VCS branch")
	if err != nil {
		return nil, err
	}

	ingressSubmodules, err := p.askOptionalBool("Ingress submodules")
	if err != nil {
		return nil, err
	}

	tagsRegex, err := p.askOptional("Tags regex")
	if err != nil {
		return nil, err
	}

	return &v1alpha1.HCPTerraformWorkspaceVCSRepoSpec{
		Identifier:        identifier,
		OAuthTokenID:      oauthTokenID,
		Branch:            branch,
		IngressSubmodules: ingressSubmodules,
		TagsRegex:         tagsRegex,
	}, nil
}

func askHCPTerraformWorkspaceVariables(p *prompter) ([]v1alpha1.HCPTerraformWorkspaceVariableSpec, error) {
	addAny, err := p.askBool("Add workspace variables", false)
	if err != nil {
		return nil, err
	}

	if !addAny {
		return nil, nil
	}

	var variables []v1alpha1.HCPTerraformWorkspaceVariableSpec
	for {
		key, err := p.askRequired("Variable key")
		if err != nil {
			return nil, err
		}

		category, err := p.askChoice("Variable category", []string{"env", "terraform"}, "env")
		if err != nil {
			return nil, err
		}

		value, err := p.askRequired("Variable value")
		if err != nil {
			return nil, err
		}

		description, err := p.askOptional("Variable description")
		if err != nil {
			return nil, err
		}

		sensitive, err := p.askOptionalBool("Sensitive variable")
		if err != nil {
			return nil, err
		}

		var hcl *bool
		if category == "terraform" {
			hcl, err = p.askOptionalBool("HCL value")
			if err != nil {
				return nil, err
			}
		}

		variables = append(variables, v1alpha1.HCPTerraformWorkspaceVariableSpec{
			Key:         key,
			Category:    category,
			Value:       value,
			Description: description,
			Sensitive:   sensitive,
			HCL:         hcl,
		})

		addAnother, err := p.askBool("Add another variable", false)
		if err != nil {
			return nil, err
		}

		if !addAnother {
			break
		}

		fmt.Fprintln(p.writer)
	}

	return variables, nil
}

func normalizeWorkspaceName(name string) string {
	return normalizeRepositoryName(name)
}
