package cli

import (
	"fmt"
	"io"
)

// Run executes the CLI against the provided arguments.
func Run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return writeHelp(stdout)
	}

	switch args[0] {
	case "help", "--help", "-h":
		return writeHelp(stdout)
	case "create-manifest":
		return runCreateManifest(args[1:], stdin, stdout)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// writeHelp renders the top-level help banner plus usage block. Colors are
// applied when stdout looks like a TTY; in plain contexts the output is
// friendly ASCII only.
func writeHelp(stdout io.Writer) error {
	s := newStyler(stdout)

	writeBanner(stdout, s)

	_, err := fmt.Fprintf(stdout, `Smyth is the authoring companion to %s. It generates validated
manifest files for reconciliation and writes them into a manifest
directory, defaulting to the current working directory.

%s
  smyth [command]

%s
  %s <type>   %s
  %s                     %s
`,
		s.bold("Anvil"),
		s.bold("Usage:"),
		s.bold("Commands:"),
		s.cyan("create-manifest"), s.dim("Author a manifest for the given resource type"),
		s.cyan("help"), s.dim("Show this message"),
	)

	return err
}
