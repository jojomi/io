package main

import "github.com/spf13/cobra"

// EnvRoot encapsulates the environment for the CLI root handler.
type EnvRoot struct {
	InputFilename    string
	TemplateFilename string
	OutputFilename   string
}

// ParseFrom reads the state from a given cobra command and its args.
func (e *EnvRoot) ParseFrom(command *cobra.Command, args []string) {
	e.InputFilename = command.Flag("input").Value.String()
	e.TemplateFilename = command.Flag("template").Value.String()
	e.OutputFilename = command.Flag("output").Value.String()
}
