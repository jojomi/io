package main

import (
	"encoding/json"
	"fmt"
	"github.com/jojomi/tplfuncs"
	"github.com/juju/errors"
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
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
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
	f.StringP("output", "o", "", "output filename including extension optionally with path")
	f.Bool("allow-exec", false, "allow execution of commands during templating phase")

	cmd.MarkFlagRequired("template")

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
	var err error

	data, err := getDataFromInput(env)
	if err != nil {
		log.Fatal().Err(err).Str("input", env.Input).Msg("failed to parse input")
	}

	// handle overwrites
	if len(env.Overwrites) > 0 {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to marshal to JSON for overwriting")
		}
		for _, overwrite := range env.Overwrites {
			key, value, ok := strings.Cut(overwrite, "=")
			if !ok {
				// invalid format, must be key=value
				continue
			}
			jsonBytes, err = sjson.SetBytes(jsonBytes, key, value)
			if err != nil {
				log.Fatal().Err(err).Str("key", key).Str("value", value).Msgf("failed to apply overwrite %s", overwrite)
			}
		}
		err = json.Unmarshal(jsonBytes, &data)
		if err != nil {
			log.Fatal().Err(err).Str("input", env.Input).Msg("failed to reunmarshal input")
		}
	}

	outputData, err := renderTemplate(env, data)
	if err != nil {
		log.Fatal().Err(err).Str("template filename", env.TemplateFilename).Msg("failed to render template")
	}

	// output to stdout?
	if env.OutputFilename == "" {
		fmt.Println(string(outputData))
		return
	}

	// output to file
	err = writeOutputFile(env, outputData, data)
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

func isHTML(env EnvRoot) bool {
	return strings.ToLower(path.Ext(env.OutputFilename)) == ".html"
}

func renderTemplate(env EnvRoot, data interface{}) ([]byte, error) {
	// read template file
	filename, err := strtpl.EvalWithFuncMap(env.TemplateFilename, getTxtFuncMap(env), data)
	if err != nil {
		return []byte{}, err
	}
	templateContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}

	var (
		result string
	)
	if isHTML(env) {
		funcMap := getHTMLFuncMap(env)
		result, err = strtpl.EvalHTMLWithFuncMap(string(templateContent), funcMap, data)
	} else {
		funcMap := getTxtFuncMap(env)
		result, err = strtpl.EvalWithFuncMap(string(templateContent), funcMap, data)
	}

	if err != nil {
		return []byte{}, err
	}
	return []byte(result), nil
}

func getHTMLFuncMap(env EnvRoot) htmlTemplate.FuncMap {
	return tplfuncs.MakeHTMLFuncMap(tplfuncs.ToHTMLFuncMap(getTxtFuncMap(env)), tplfuncs.HTMLSafeHelpers())
}

func getTxtFuncMap(env EnvRoot) template.FuncMap {
	maps := []template.FuncMap{
		sprig.TxtFuncMap(),
		tplfuncs.SpacingHelpers(),
		tplfuncs.LineHelpers(),
	}
	if env.AllowExec {
		maps = append(maps, tplfuncs.ExecHelpers())
	}
	return tplfuncs.MakeFuncMap(maps...)
}

func writeOutputFile(env EnvRoot, content []byte, data interface{}) error {
	// eval template on OutputFilename
	filename, err := strtpl.EvalWithFuncMap(env.OutputFilename, getTxtFuncMap(env), data)
	if err != nil {
		return errors.Annotatef(err, "could not evaluate output filename template: %s", env.OutputFilename)
	}

	// make sure the target dir exists
	err = os.MkdirAll(path.Dir(filename), os.FileMode(0750))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, content, os.FileMode(0640))
	if err != nil {
		return err
	}

	return nil
}
