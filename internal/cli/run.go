package cli

import (
	"fmt"
	"io"
)

const helpText = `Hello from Smyth.

This is the current Smyth CLI scaffold.

Smyth is the authoring companion to Anvil. It will generate validated
manifest files for reconciliation and write them into a manifest directory,
defaulting to the current working directory.

Usage:
  smyth [command]

Available Commands:
  help        Show this message
`

// Run executes the CLI against the provided arguments.
func Run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		_, err := fmt.Fprint(stdout, helpText)
		return err
	}

	switch args[0] {
	case "help", "--help", "-h":
		_, err := fmt.Fprint(stdout, helpText)
		return err
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
