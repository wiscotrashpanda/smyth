package cli

import (
	"fmt"
	"io"
)

const createManifestHelp = `Usage:
  smyth create-manifest <type> [flags]

Authors a YAML manifest for the requested resource type. Each type has its own
prompt flow and shared-schema validation.

Manifest types:
  github-repo   Author a GitHubRepository manifest

Run ` + "`smyth create-manifest <type> --help`" + ` for type-specific flags.
`

// runCreateManifest dispatches to the per-type authoring flow. Keeping this
// layer separate from Run makes it obvious that `create-manifest` produces a
// manifest file; the type argument decides which schema drives the prompts.
func runCreateManifest(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		_, err := fmt.Fprint(stdout, createManifestHelp)
		return err
	}

	switch args[0] {
	case "help", "--help", "-h":
		_, err := fmt.Fprint(stdout, createManifestHelp)
		return err
	case "github-repo":
		return runCreateGitHubRepository(args[1:], stdin, stdout)
	default:
		return fmt.Errorf("unknown manifest type: %s", args[0])
	}
}
