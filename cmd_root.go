package main

import (
	"encoding/json"
	"fmt"
	"github.com/jojomi/tplfuncs"
	htmlTemplate "html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

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

	f := cmd.Flags()
	f.StringP("input", "i", "", "input filename including extension optionally with path, or inline JSON if first char is {")
	f.StringP("template", "t", "", "template filename including extension optionally with path")
	f.StringP("output", "o", "", "output filename including extension optionally with path")
	f.Bool("allow-exec", false, "allow execution of commands during templating phase")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("template")
	cmd.MarkFlagRequired("output")

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
	data, err := getDataFromInput(env)
	if err != nil {
		log.Fatal().Err(err).Str("input", env.Input).Msg("failed to parse input")
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

	// inline JSON?
	if len(env.Input) > 0 && env.Input[0:1] == "{" {
		var data map[string]interface{}
		err := json.Unmarshal([]byte(env.Input), &data)
		if err != nil {
			return nil, fmt.Errorf("invalid inline json: %s", env.Input)
		}
		return data, nil
	}

	switch strings.ToLower(path.Ext(env.Input)) {
	case ".yml":
		fallthrough
	case ".yaml":
		data, err = input.GetYamlFromFile(env.Input)
	case ".json":
		data, err = input.GetJSONFromFile(env.Input)
	case ".csv":
		data, err = input.GetCSVFromFile(env.Input)
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

	var (
		result string
	)
	switch strings.ToLower(path.Ext(env.OutputFilename)) {
	case ".html":
		maps := []htmlTemplate.FuncMap{
			sprig.FuncMap(),
			tplfuncs.SpacingHelpersHTML(),
			tplfuncs.LineHelpersHTML(),
		}
		if env.AllowExec {
			maps = append(maps, tplfuncs.ExecHelpersHTML())
		}
		funcMap := tplfuncs.MakeHTMLFuncMap(maps...)
		result, err = strtpl.EvalHTMLWithFuncMap(string(templateContent), funcMap, data)
	default:
		maps := []template.FuncMap{
			sprig.TxtFuncMap(),
			tplfuncs.SpacingHelpers(),
			tplfuncs.LineHelpers(),
		}
		if env.AllowExec {
			maps = append(maps, tplfuncs.ExecHelpers())
		}
		funcMap := tplfuncs.MakeFuncMap(maps...)
		result, err = strtpl.EvalWithFuncMap(string(templateContent), funcMap, data)
	}

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
