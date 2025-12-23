package io

import "text/template"

// IOOpts encapsulates the options for IO.
type IOOpts struct {
	Data       any
	DataFile   string
	DataString string

	Overwrites []string

	TemplateFilename string
	TemplateInline   string

	OutputFilename string

	CustomFuncMap *template.FuncMap

	AllowExec    bool
	AllowNetwork bool
	AllowIO      bool
}
