package io

import "text/template"

// IOOpts encapsulates the options for IO.
type IOOpts struct {
	Input            string
	Overwrites       []string
	TemplateFilename string
	TemplateInline   string
	OutputFilename   string
	AllowExec        bool
	AllowNetwork     bool
	AllowIO          bool
	CustomFuncMap    *template.FuncMap
}
