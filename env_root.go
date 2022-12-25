package main

import "github.com/spf13/cobra"

// EnvRoot encapsulates the environment for the CLI root handler.
type EnvRoot struct {
	Input            string
	Overwrites       []string
	TemplateFilename string
	TemplateInline   string
	OutputFilename   string
	AllowExec        bool
	AllowNetwork     bool
	AllowIO          bool
}

// ParseFrom reads the state from a given cobra command and its args.
func (e *EnvRoot) ParseFrom(command *cobra.Command, args []string) error {
	var (
		f   = command.Flags()
		err error
	)

	e.Input, err = f.GetString("input")
	e.TemplateFilename = command.Flag("template").Value.String()
	e.TemplateInline = command.Flag("template-inline").Value.String()
	e.OutputFilename = command.Flag("output").Value.String()

	e.Overwrites, err = f.GetStringArray("overwrite")
	if err != nil {
		return err
	}
	e.AllowExec, err = command.Flags().GetBool("allow-exec")
	if err != nil {
		return err
	}
	e.AllowIO, err = command.Flags().GetBool("allow-io")
	if err != nil {
		return err
	}
	e.AllowNetwork, err = command.Flags().GetBool("allow-network")
	if err != nil {
		return err
	}

	return nil
}
