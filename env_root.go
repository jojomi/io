package main

import "github.com/spf13/cobra"

// EnvRoot encapsulates the environment for the CLI root handler.
type EnvRoot struct {
	InputFilename    string
	TemplateFilename string
	OutputFilename   string
	AllowExec        bool
}

// ParseFrom reads the state from a given cobra command and its args.
func (e *EnvRoot) ParseFrom(command *cobra.Command, args []string) error {
	e.InputFilename = command.Flag("input").Value.String()
	e.TemplateFilename = command.Flag("template").Value.String()
	e.OutputFilename = command.Flag("output").Value.String()

	var err error
	e.AllowExec, err = command.Flags().GetBool("allow-exec")
	if err != nil {
		return err
	}
	return nil
}
