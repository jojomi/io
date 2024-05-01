package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlToHTML(t *testing.T) {
	env := EnvRoot{
		Input:            "test/input/simple.yml",
		TemplateFilename: "test/template/creator.html",
		OutputFilename:   "test/output/test.html",
	}
	t.Cleanup(func() {
		os.Remove(env.OutputFilename)
	})

	handleRoot(env)

	outputContent, err := os.ReadFile(env.OutputFilename)
	assert.FileExists(t, env.OutputFilename)
	assert.NoError(t, err)
	assert.Contains(t, string(outputContent), "54", "The output contains the age.")
	assert.Contains(t, string(outputContent), "John Doe", "The output contains the name.")
}

func TestJSONToHTML(t *testing.T) {
	env := EnvRoot{
		Input:            "test/input/simple.json",
		TemplateFilename: "test/template/creator.html",
		OutputFilename:   "test/output/test.html",
	}
	t.Cleanup(func() {
		os.Remove(env.OutputFilename)
	})

	handleRoot(env)

	outputContent, err := os.ReadFile(env.OutputFilename)
	assert.FileExists(t, env.OutputFilename)
	assert.NoError(t, err)
	assert.Contains(t, string(outputContent), "54", "The output contains the age.")
	assert.Contains(t, string(outputContent), "John Doe", "The output contains the name.")
}

func TestCSVToHTML(t *testing.T) {
	env := EnvRoot{
		Input:            "test/input/simple.csv",
		TemplateFilename: "test/template/creator_csv.html",
		OutputFilename:   "test/output/test.html",
	}
	t.Cleanup(func() {
		os.Remove(env.OutputFilename)
	})

	handleRoot(env)

	outputContent, err := os.ReadFile(env.OutputFilename)
	assert.FileExists(t, env.OutputFilename)
	assert.NoError(t, err)
	assert.Contains(t, string(outputContent), "54", "The output contains the age.")
	assert.Contains(t, string(outputContent), "John Doe", "The output contains the name.")
}

func TestJSONToHTMLInline(t *testing.T) {
	env := EnvRoot{
		Input:            `{"creator": { "name": "John Doe", "age": 54} }`,
		TemplateFilename: `test/template/creat {{- if eq .creator.name "John Doe" -}} or {{- end -}} .html`,
		OutputFilename:   "test/output/test-{{ .creator.age }}.html",
	}
	t.Cleanup(func() {
		os.Remove(env.OutputFilename)
	})

	handleRoot(env)

	effectiveFilename := "test/output/test-54.html"
	assert.FileExists(t, effectiveFilename)
	outputContent, err := os.ReadFile(effectiveFilename)
	assert.NoError(t, err)
	assert.Contains(t, string(outputContent), "54", "The output contains the age.")
	assert.Contains(t, string(outputContent), "John Doe", "The output contains the name.")
}
