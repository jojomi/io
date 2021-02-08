package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlToHTML(t *testing.T) {
	env := EnvRoot{
		InputFilename:    "test/input/simple.yml",
		TemplateFilename: "test/template/creator.html",
		OutputFilename:   "test/output/test.html",
	}
	t.Cleanup(func() {
		os.Remove(env.OutputFilename)
	})

	handleRoot(env)

	outputContent, err := ioutil.ReadFile(env.OutputFilename)
	assert.FileExists(t, env.OutputFilename)
	assert.NoError(t, err)
	assert.Contains(t, string(outputContent), "54", "The output contains the age.")
	assert.Contains(t, string(outputContent), "John Doe", "The output contains the name.")
}

func TestCSVToHTML(t *testing.T) {
	env := EnvRoot{
		InputFilename:    "test/input/simple.csv",
		TemplateFilename: "test/template/creator_csv.html",
		OutputFilename:   "test/output/test.html",
	}
	t.Cleanup(func() {
		os.Remove(env.OutputFilename)
	})

	handleRoot(env)

	outputContent, err := ioutil.ReadFile(env.OutputFilename)
	assert.FileExists(t, env.OutputFilename)
	assert.NoError(t, err)
	assert.Contains(t, string(outputContent), "54", "The output contains the age.")
	assert.Contains(t, string(outputContent), "John Doe", "The output contains the name.")
}
