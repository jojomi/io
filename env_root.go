package main

import "github.com/spf13/cobra"

// EnvRoot encapsulates the environment for the CLI root handler.
type EnvRoot struct {
}

// ParseFrom reads the state from a given cobra command and its args.
func (e *EnvRoot) ParseFrom(command *cobra.Command, args []string) {

}
