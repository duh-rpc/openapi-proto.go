package internal

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const protoTemplate = `syntax = "proto3";

package {{.PackageName}};
{{range .Messages}}
{{if .Description}}{{formatComment .Description}}{{end}}message {{.Name}} {
{{range .Fields}}{{if .Description}}  {{formatComment .Description}}{{end}}  {{if .Repeated}}repeated {{end}}{{.Type}} {{.Name}} = {{.Number}}{{if .JSONName}} [json_name = "{{.JSONName}}"]{{end}};
{{end}}}
{{end}}`

type templateData struct {
	PackageName string
	Messages    []*ProtoMessage
	Enums       []*ProtoEnum
}

// Generate creates proto3 output from messages and enums
func Generate(packageName string, messages []*ProtoMessage, enums []*ProtoEnum) ([]byte, error) {
	funcMap := template.FuncMap{
		"formatComment": formatCommentForTemplate,
	}

	tmpl, err := template.New("proto").Funcs(funcMap).Parse(protoTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	data := templateData{
		PackageName: packageName,
		Messages:    messages,
		Enums:       enums,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// formatCommentForTemplate formats a description as a proto3 comment for use in templates
func formatCommentForTemplate(description string) string {
	if strings.TrimSpace(description) == "" {
		return ""
	}

	lines := strings.Split(description, "\n")
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed == "" {
			result.WriteString("//\n")
		} else {
			result.WriteString("// ")
			result.WriteString(trimmed)
			result.WriteString("\n")
		}
	}

	return result.String()
}
