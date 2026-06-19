package generator

import (
	"bytes"
	"text/template"
)

// parseTemplateWithFuncs parses a template string with custom functions.
func parseTemplateWithFuncs(name, tmpl string, funcs template.FuncMap) (*template.Template, error) {
	return template.New(name).Funcs(funcs).Parse(tmpl)
}

// execTemplate executes a pre-parsed template with the given data.
func execTemplate(t *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
