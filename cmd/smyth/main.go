package main

import (
	"fmt"
	"os"

	"github.com/wiscotrashpanda/smyth/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
