package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/iancoleman/strcase"
	"github.com/jojomi/io/input"
	"github.com/jojomi/strtpl"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func getRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: strcase.ToKebab(ToolName),
		Run: handleRootCmd,
	}

	f := cmd.PersistentFlags()
	f.StringP("input", "i", "", "input filename including extension optionally with path")
	f.StringP("template", "t", "", "template filename including extension optionally with path")
	f.StringP("output", "o", "", "output filename including extension optionally with path")

	cmd.MarkPersistentFlagRequired("input")
	cmd.MarkPersistentFlagRequired("template")
	cmd.MarkPersistentFlagRequired("output")

	return cmd
}

func handleRootCmd(cmd *cobra.Command, args []string) {
	env := EnvRoot{}
	env.ParseFrom(cmd, args)
	handleRoot(env)
}

func handleRoot(env EnvRoot) {
	data, err := getDataFromInput(env)
	if err != nil {
		log.Fatal().Err(err).Str("input filename", env.InputFilename).Msg("failed to parse input file")
	}

	outputData, err := renderTemplate(env, data)
	if err != nil {
		log.Fatal().Err(err).Str("template filename", env.TemplateFilename).Msg("failed to render template")
	}

	err = writeOutputFile(env, outputData)
	if err != nil {
		log.Fatal().Err(err).Str("output filename", env.OutputFilename).Msg("failed to write to output file")
	}
}

func getDataFromInput(env EnvRoot) (interface{}, error) {
	var (
		data interface{}
		err  error
	)

	switch strings.ToLower(path.Ext(env.InputFilename)) {
	case ".yml":
		fallthrough
	case ".yaml":
		data, err = input.GetYamlFromFile(env.InputFilename)
	case ".csv":
		data, err = input.GetCSVFromFile(env.InputFilename)
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

func renderTemplate(env EnvRoot, data interface{}) ([]byte, error) {
	// read template file
	templateContent, err := ioutil.ReadFile(env.TemplateFilename)
	if err != nil {
		return []byte{}, err
	}

	result, err := strtpl.EvalWithFuncMap(string(templateContent), sprig.TxtFuncMap(), data)
	if err != nil {
		return []byte{}, err
	}
	return []byte(result), nil
}

func writeOutputFile(env EnvRoot, content []byte) error {
	// make sure the target dir exists
	err := os.MkdirAll(path.Dir(env.OutputFilename), os.FileMode(0750))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(env.OutputFilename, content, os.FileMode(0640))
	if err != nil {
		return err
	}

	return nil
}
