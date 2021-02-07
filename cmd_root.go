package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
)

func getRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: strcase.ToKebab(ToolName),
	}
	return cmd
}

func handleRootCmd(cmd *cobra.Command, args []string) {
	env := EnvRoot{}
	env.ParseFrom(cmd, args)
	handleRoot(env)
}

func handleRoot(env EnvRoot) {
	// implement logic here

	// access config like this
	conf := mustGetConfig(configFilename)
	fmt.Println(conf)
}
