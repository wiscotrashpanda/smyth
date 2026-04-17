package cli

import (
	"fmt"
	"io"
)

// runCreateManifest dispatches to the per-type authoring flow. Keeping this
// layer separate from Run makes it obvious that `create-manifest` produces a
// manifest file; the type argument decides which schema drives the prompts.
func runCreateManifest(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return writeCreateManifestHelp(stdout)
	}

	switch args[0] {
	case "help", "--help", "-h":
		return writeCreateManifestHelp(stdout)
	case "github-repo":
		return runCreateGitHubRepository(args[1:], stdin, stdout)
	case "hcp-terraform-workspace":
		return runCreateHCPTerraformWorkspace(args[1:], stdin, stdout)
	default:
		return fmt.Errorf("unknown manifest type: %s", args[0])
	}
}

func writeCreateManifestHelp(stdout io.Writer) error {
	s := newStyler(stdout)

	_, err := fmt.Fprintf(stdout, `%s
  smyth create-manifest <type> [flags]

Authors a YAML manifest for the requested resource type. Each type has its own
prompt flow and shared-schema validation.

%s
  %s   %s
  %s   %s

Run %s for type-specific flags.
`,
		s.bold("Usage:"),
		s.bold("Manifest types:"),
		s.cyan("github-repo"), s.dim("Author a GitHubRepository manifest"),
		s.cyan("hcp-terraform-workspace"), s.dim("Author a HCPTerraformWorkspace manifest"),
		s.cyan("smyth create-manifest <type> --help"),
	)

	return err
}
