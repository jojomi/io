package io

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

	"github.com/jojomi/io/input"
	"github.com/jojomi/strtpl"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/sjson"
)

func RenderFile(opts IOOpts) error {
	data, outputData, err := generateOutput(opts)
	if err != nil {
		log.Fatal().Err(err).Interface("opts", opts).Interface("data", data).Msg("failed to run")
	}

	// output to file
	err = writeOutputFile(opts, []byte(outputData), data)
	if err != nil {
		log.Fatal().Err(err).Str("output filename", opts.OutputFilename).Msg("failed to write to output file")
	}

	return nil
}

func RenderString(opts IOOpts) (string, error) {
	data, outputData, err := generateOutput(opts)
	if err != nil {
		log.Fatal().Err(err).Interface("opts", opts).Interface("data", data).Msg("failed to run")
	}
	return string(outputData), err
}

// generateOutput generates the output content using the environment given. This is the workhorse inside io.
// It returns the inputData used and the template output bytes plus the error if one occurred.
func generateOutput(opts IOOpts) (interface{}, []byte, error) {
	inputData, err := getDataFromInput(opts)
	if err != nil {
		return nil, nil, errors.Annotatef(err, "failed to parse input %s", opts.Input)
	}

	templateContent, err := getTemplateContent(opts, inputData)
	if err != nil {
		return nil, []byte{}, err
	}

	return generateOutputForTemplate(opts, inputData, templateContent)
}

func generateOutputForTemplate(opts IOOpts, inputData interface{}, templateData []byte) (interface{}, []byte, error) {
	// handle overwrites
	if len(opts.Overwrites) > 0 {
		var jsonEncoder = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := jsonEncoder.Marshal(inputData)
		if err != nil {
			return nil, nil, errors.Annotate(err, "failed to marshal to JSON for overwriting")
		}
		for _, overwrite := range opts.Overwrites {
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
			return nil, nil, errors.Annotatef(err, "failed to re-unmarshal input %s", opts.Input)
		}
	}

	outputData, err := renderTemplate(opts, inputData, templateData)
	if err != nil {
		return nil, nil, errors.Annotatef(err, "failed to render template %s", opts.TemplateFilename)
	}
	return inputData, outputData, nil
}

func getDataFromInput(opts IOOpts) (interface{}, error) {
	var (
		data interface{}
		err  error
	)

	// inline JSON?
	if len(opts.Input) > 0 && opts.Input[0:1] == "{" {
		var data map[string]interface{}
		err := json.Unmarshal([]byte(opts.Input), &data)
		if err != nil {
			return nil, fmt.Errorf("invalid inline json: %s", opts.Input)
		}
		return data, nil
	}

	switch strings.ToLower(path.Ext(opts.Input)) {
	case ".yml":
		fallthrough
	case ".yaml":
		data, err = input.GetYamlFromFile(opts.Input)
	case ".json":
		data, err = input.GetJSONFromFile(opts.Input)
	case ".csv":
		data, err = input.GetCSVFromFile(opts.Input)
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

func isHTML(opts IOOpts) bool {
	return strings.ToLower(path.Ext(opts.OutputFilename)) == ".html"
}

func getTemplateContent(opts IOOpts, data interface{}) ([]byte, error) {
	// validate flags
	if opts.TemplateFilename != "" && opts.TemplateInline != "" {
		return []byte{}, fmt.Errorf("both --template and --template-inline are set, pick one")
	}
	if opts.TemplateFilename == "" && opts.TemplateInline == "" {
		return []byte{}, fmt.Errorf("neither --template nor --template-inline are set, set one")
	}

	// read template file
	if opts.TemplateFilename != "" {
		filename, err := strtpl.EvalWithFuncMap(opts.TemplateFilename, getTxtFuncMap(opts), data)
		if err != nil {
			return []byte{}, err
		}
		return os.ReadFile(filename)
	}

	return []byte(opts.TemplateInline), nil
}

func renderTemplate(opts IOOpts, data interface{}, templateData []byte) ([]byte, error) {
	var (
		result string
		err    error
	)
	if isHTML(opts) {
		funcMap := getHTMLFuncMap(opts)
		result, err = strtpl.EvalHTMLWithFuncMap(string(templateData), funcMap, data)
	} else {
		funcMap := getTxtFuncMap(opts)
		result, err = strtpl.EvalWithFuncMap(string(templateData), funcMap, data)
	}
	if err != nil {
		return []byte{}, err
	}

	return []byte(result), nil
}

func getHTMLFuncMap(opts IOOpts) htmlTemplate.FuncMap {
	return tplfuncs.MakeHTMLFuncMap(tplfuncs.ToHTMLFuncMap(getTxtFuncMap(opts)), tplfuncs.HTMLSafeHelpers())
}

func getTxtFuncMap(opts IOOpts) template.FuncMap {
	maps := []template.FuncMap{
		tplfuncs.AssertHelpers(),
		tplfuncs.CastHelpers(),
		tplfuncs.ContainerHelpers(),
		tplfuncs.DateHelpers(),
		tplfuncs.DefaultHelpers(),
		tplfuncs.EncodeHelpers(),
		tplfuncs.GolangHelpers(),
		tplfuncs.HashHelpers(),
		tplfuncs.JSONHelpers(),
		tplfuncs.LanguageHelpers(),
		tplfuncs.LinesHelpers(),
		tplfuncs.LoopHelpers(),
		tplfuncs.MathHelpers(),
		tplfuncs.PrintHelpers(),
		tplfuncs.RandomHelpers(),
		tplfuncs.SemverHelpers(),
		tplfuncs.SpacingHelpers(),
		tplfuncs.StringHelpers(),
		tplfuncs.TypeConversionHelpers(),
		tplfuncs.YAMLHelpers(),
	}
	if opts.AllowExec {
		maps = append(
			maps,
			tplfuncs.ExecHelpers(),
		)
	}
	if opts.AllowIO || opts.AllowExec {
		maps = append(
			maps,
			tplfuncs.IOHelpers(),
			tplfuncs.EnvHelpers(),
			tplfuncs.FilesystemHelpers(),
		)
	}
	if opts.AllowNetwork || opts.AllowExec {
		maps = append(
			maps,
			tplfuncs.NetworkHelpers(),
		)
	}

	if opts.CustomFuncMap != nil {
		maps = append(maps, *opts.CustomFuncMap)
	}

	result := tplfuncs.MakeFuncMap(maps...)

	// io aware include function (with same data)
	inlineTemplateFunc := func(filename string) (string, error) {
		inputData, err := getDataFromInput(opts)
		if err != nil {
			return "", errors.Annotatef(err, "failed to parse input %s", opts.Input)
		}

		templateData, err := os.ReadFile(filename)
		if err != nil {
			return "", errors.Annotatef(err, "failed to read include file at %s", filename)
		}

		_, out, err := generateOutputForTemplate(opts, inputData, templateData)
		return string(out), errors.Annotatef(err, "failed to render inlined template %s (no data)", filename)
	}

	inlineWithDataFunc := func(filename string, data ...interface{}) (string, error) {
		var (
			inputData interface{}
			err       error
		)

		if len(data) == 1 {
			inputData = data[0]
		} else {
			inputData, err = getMapFromParams(data)
			if err != nil {
				return "", errors.Annotatef(err, "failed to parse input from %v", data)
			}
		}

		templateData, err := os.ReadFile(filename)
		if err != nil {
			return "", errors.Annotatef(err, "failed to read include file at %s", filename)
		}

		_, out, err := generateOutputForTemplate(opts, inputData, templateData)
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

func writeOutputFile(opts IOOpts, content []byte, data interface{}) error {
	// no filename given -> stdout!
	if opts.OutputFilename == "" {
		fmt.Println(string(content))
		return nil
	}

	// eval template on OutputFilename
	filename, err := strtpl.EvalWithFuncMap(opts.OutputFilename, getTxtFuncMap(opts), data)
	if err != nil {
		return errors.Annotatef(err, "could not evaluate output filename template: %s", opts.OutputFilename)
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
