package main

import (
	"github.com/iancoleman/strcase"
	jio "github.com/jojomi/io"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func getRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: strcase.ToKebab(ToolName),
		Run: handleRootCmd,
	}

	f := cmd.Flags()
	f.StringP("input", "i", "{}", "input filename including extension optionally with path, or inline JSON if first char is {")
	f.StringArrayP("overwrite", "w", []string{}, "overwrite input data by path (for YML and JSON inputs only)")
	f.StringP("template", "t", "", "template filename including extension optionally with path")
	f.String("template-inline", "", "inline template content")
	f.StringP("output", "o", "", "output filename including extension optionally with path")
	f.Bool("allow-exec", false, "allow execution of commands during templating phase, implies --allow-io and --allow-network")
	f.Bool("allow-io", false, "allow reading and writing files during templating phase")
	f.Bool("allow-network", false, "allow network communication during templating phase")

	cmd.MarkFlagsMutuallyExclusive("template", "template-inline")

	return cmd
}

func handleRootCmd(cmd *cobra.Command, args []string) {
	env := EnvRoot{}
	err := env.ParseFrom(cmd, args)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse params")
	}
	handleRoot(env)
}

func handleRoot(env EnvRoot) {
	opts := jio.IOOpts{
		Input:            env.Input,
		Overwrites:       env.Overwrites,
		TemplateFilename: env.TemplateFilename,
		TemplateInline:   env.TemplateInline,
		OutputFilename:   env.OutputFilename,
		AllowExec:        env.AllowExec,
		AllowNetwork:     env.AllowNetwork,
		AllowIO:          env.AllowIO,
	}
	_ = opts
	err := jio.RenderFile(opts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to render")
	}
	return
}
