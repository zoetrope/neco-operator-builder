package cmd

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Main{}

// Main scaffolds main.go
type Main struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ProjectNameMixin
}

// SetTemplateDefaults implements file.Template
func (f *Main) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("cmd", f.ProjectName+"-controller", "main.go")
	}

	f.TemplateBody = mainTemplate

	return nil
}

//nolint:lll
const mainTemplate = `{{ .Boilerplate }}

package main

import "{{ .Repo }}/cmd/{{ .ProjectName }}-controller/sub"

func main() {
	sub.Execute()
}
`
