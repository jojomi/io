package main

import (
	"encoding/json"
	"fmt"
	htmlTemplate "html/template"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/jojomi/tplfuncs"
	"github.com/juju/errors"

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
	data, outputData, err := generateOutput(env)
	if err != nil {
		log.Fatal().Err(err).Interface("env", env).Msg("failed to run")
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

// generateOutput generates the output content using the environment given. This is the workhorse inside io.
func generateOutput(env EnvRoot) (interface{}, []byte, error) {
	inputData, err := getDataFromInput(env)
	if err != nil {
		return nil, nil, errors.Annotatef(err, "failed to parse input %s", env.Input)
	}

	templateContent, err := getTemplateContent(env, inputData)
	if err != nil {
		return nil, []byte{}, err
	}

	return generateOutputForTemplate(env, inputData, templateContent)
}

func generateOutputForTemplate(env EnvRoot, inputData interface{}, templateData []byte) (interface{}, []byte, error) {
	// handle overwrites
	if len(env.Overwrites) > 0 {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := json.Marshal(inputData)
		if err != nil {
			return nil, nil, errors.Annotate(err, "failed to marshal to JSON for overwriting")
		}
		for _, overwrite := range env.Overwrites {
			key, value, ok := strings.Cut(overwrite, "=")
			if !ok {
				// invalid format, must be key=value
				continue
			}
			jsonBytes, err = sjson.SetBytes(jsonBytes, key, value)
			if err != nil {
				return nil, nil, errors.Annotatef(err, "failed to apply overwrite %s (key=%s, value=%s)", overwrite, key, value)
			}
		}
		err = json.Unmarshal(jsonBytes, &inputData)
		if err != nil {
			return nil, nil, errors.Annotatef(err, "failed to re-unmarshal input %s", env.Input)
		}
	}

	outputData, err := renderTemplate(env, inputData, templateData)
	if err != nil {
		return nil, nil, errors.Annotatef(err, "failed to render template %s", env.TemplateFilename)
	}
	return inputData, outputData, nil
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

func getTemplateContent(env EnvRoot, data interface{}) ([]byte, error) {
	// validate flags
	if env.TemplateFilename != "" && env.TemplateInline != "" {
		return []byte{}, fmt.Errorf("both --template and --template-inline are set, pick one")
	}
	if env.TemplateFilename == "" && env.TemplateInline == "" {
		return []byte{}, fmt.Errorf("neither --template nor --template-inline are set, set one")
	}

	// read template file
	if env.TemplateFilename != "" {
		filename, err := strtpl.EvalWithFuncMap(env.TemplateFilename, getTxtFuncMap(env), data)
		if err != nil {
			return []byte{}, err
		}
		return os.ReadFile(filename)
	}

	return []byte(env.TemplateInline), nil
}

func renderTemplate(env EnvRoot, data interface{}, templateData []byte) ([]byte, error) {
	var (
		result string
		err    error
	)
	if isHTML(env) {
		funcMap := getHTMLFuncMap(env)
		result, err = strtpl.EvalHTMLWithFuncMap(string(templateData), funcMap, data)
	} else {
		funcMap := getTxtFuncMap(env)
		result, err = strtpl.EvalWithFuncMap(string(templateData), funcMap, data)
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
		/// tplfuncs.CastHelpers(),
		tplfuncs.SpacingHelpers(),
		tplfuncs.OutputHelpers(),
		tplfuncs.ContainerHelpers(),
		tplfuncs.StringHelpers(),
		tplfuncs.MathHelpers(),
		tplfuncs.JSONHelpers(),
		tplfuncs.LineHelpers(),
		tplfuncs.EnvHelpers(),
		tplfuncs.FilesystemHelpers(),
		tplfuncs.LanguageHelpers(),
		tplfuncs.HashHelpers(),
		tplfuncs.SemverHelpers(),
	}
	if env.AllowExec {
		maps = append(maps, tplfuncs.ExecHelpers())
	}
	if env.AllowIO || env.AllowExec {
		maps = append(maps, tplfuncs.IOHelpers())
	}
	if env.AllowNetwork || env.AllowExec {
		maps = append(maps, tplfuncs.NetworkHelpers())
	}

	result := tplfuncs.MakeFuncMap(maps...)

	// io aware include function (with same data)
	inlineTemplateFunc := func(filename string) (string, error) {
		inputData, err := getDataFromInput(env)
		if err != nil {
			return "", errors.Annotatef(err, "failed to parse input %s", env.Input)
		}

		templateData, err := os.ReadFile(filename)
		if err != nil {
			return "", errors.Annotatef(err, "failed to read include file at %s", filename)
		}

		_, out, err := generateOutputForTemplate(env, inputData, templateData)
		return string(out), errors.Annotatef(err, "failed to render inlined template %s (no data)", filename)
	}

	inlineWithDataFunc := func(filename string, data ...interface{}) (string, error) {
		inputData, err := getMapFromParams(data)
		if err != nil {
			return "", errors.Annotatef(err, "failed to parse input from %v", data)
		}

		templateData, err := os.ReadFile(filename)
		if err != nil {
			return "", errors.Annotatef(err, "failed to read include file at %s", filename)
		}

		_, out, err := generateOutputForTemplate(env, inputData, templateData)
		return string(out), errors.Annotatef(err, "failed to render inlined template %s with data %v", filename, inputData)
	}

	result["inline"] = inlineTemplateFunc
	result["includeIO"] = inlineTemplateFunc // for backward-compatibility only!

	result["inlineIfExists"] = func(filename string) (string, error) {
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return inlineTemplateFunc(filename)
	}

	// TODO result["inlineRaw"] = ?

	// add io aware include function (with explicitly given data)
	result["inlineWithData"] = inlineWithDataFunc

	result["inlineIfExistsWithData"] = func(filename string, data ...interface{}) (string, error) {
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return inlineWithDataFunc(filename, data...)
	}

	return result
}

func getMapFromParams(data []interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var err error
	for i := 0; i < len(data)-1; i = i + 2 {
		if key, ok := data[i].(string); ok {
			result[key] = data[i+1]
			continue
		}
		if err == nil {
			err = fmt.Errorf("could not parse %v for config key (needs to be string)", data[i])
		}
	}

	return result, err
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

	err = os.WriteFile(filename, content, os.FileMode(0640))
	if err != nil {
		return err
	}

	return nil
}
