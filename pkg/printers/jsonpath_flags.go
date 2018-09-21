/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package printers

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

// JSONPathPrintFlags provides default flags necessary for template printing.
// Given the following flag values, a printer can be requested that knows
// how to handle printing based on these values.
type JSONPathPrintFlags struct {
	// indicates if it is OK to ignore missing keys for rendering
	// an output template.
	AllowMissingKeys *bool
	TemplateArgument *string
}

// ToPrinter receives an templateFormat and returns a printer capable of
// handling --template format printing.
// Returns false if the specified templateFormat does not match a template format.
func (f *JSONPathPrintFlags) ToPrinter(templateFormat string) (ResourcePrinter, error) {
	if (f.TemplateArgument == nil || len(*f.TemplateArgument) == 0) && len(templateFormat) == 0 {
		return nil, genericclioptions.NoCompatiblePrinterError{Options: f, OutputFormat: &templateFormat}
	}

	templateValue := ""

	// templates are logically optional for specifying a format.
	// this allows a user to specify a template format value
	// as --output=jsonpath=
	templateFormats := map[string]bool{
		"jsonpath":      true,
		"jsonpath-file": true,
	}

	if f.TemplateArgument == nil || len(*f.TemplateArgument) == 0 {
		for format := range templateFormats {
			format = format + "="
			if strings.HasPrefix(templateFormat, format) {
				templateValue = templateFormat[len(format):]
				templateFormat = format[:len(format)-1]
				break
			}
		}
	} else {
		templateValue = *f.TemplateArgument
	}

	if _, supportedFormat := templateFormats[templateFormat]; !supportedFormat {
		return nil, genericclioptions.NoCompatiblePrinterError{Options: f, OutputFormat: &templateFormat}
	}

	if len(templateValue) == 0 {
		return nil, fmt.Errorf("template format specified but no template given")
	}

	if templateFormat == "jsonpath-file" {
		data, err := ioutil.ReadFile(templateValue)
		if err != nil {
			return nil, fmt.Errorf("error reading --template %s, %v\n", templateValue, err)
		}

		templateValue = string(data)
	}

	p, err := NewJSONPathPrinter(templateValue)
	if err != nil {
		return nil, fmt.Errorf("error parsing jsonpath %s, %v\n", templateValue, err)
	}

	allowMissingKeys := true
	if f.AllowMissingKeys != nil {
		allowMissingKeys = *f.AllowMissingKeys
	}

	p.AllowMissingKeys(allowMissingKeys)
	return p, nil
}

// AddFlags receives a *cobra.Command reference and binds
// flags related to template printing to it
func (f *JSONPathPrintFlags) AddFlags(c *cobra.Command) {
	if f.TemplateArgument != nil {
		c.Flags().StringVar(f.TemplateArgument, "template", *f.TemplateArgument, "Template string or path to template file to use when --output=jsonpath, --output=jsonpath-file.")
		c.MarkFlagFilename("template")
	}
	if f.AllowMissingKeys != nil {
		c.Flags().BoolVar(f.AllowMissingKeys, "allow-missing-template-keys", *f.AllowMissingKeys, "If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats.")
	}
}

// NewJSONPathPrintFlags returns flags associated with
// --template printing, with default values set.
func NewJSONPathPrintFlags(templateValue string, allowMissingKeys bool) *JSONPathPrintFlags {
	return &JSONPathPrintFlags{
		TemplateArgument: &templateValue,
		AllowMissingKeys: &allowMissingKeys,
	}
}
