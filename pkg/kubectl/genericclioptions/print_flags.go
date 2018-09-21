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

package genericclioptions

import (
	"fmt"
<<<<<<< HEAD
	"sort"
	"strings"

	"github.com/spf13/cobra"

=======

	"github.com/spf13/cobra"
>>>>>>> axbaretto
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions/printers"
)

type NoCompatiblePrinterError struct {
<<<<<<< HEAD
	OutputFormat   *string
	AllowedFormats []string
	Options        interface{}
=======
	OutputFormat *string
	Options      interface{}
>>>>>>> axbaretto
}

func (e NoCompatiblePrinterError) Error() string {
	output := ""
	if e.OutputFormat != nil {
		output = *e.OutputFormat
	}

<<<<<<< HEAD
	sort.Strings(e.AllowedFormats)
	return fmt.Sprintf("unable to match a printer suitable for the output format %q, allowed formats are: %s", output, strings.Join(e.AllowedFormats, ","))
=======
	return fmt.Sprintf("unable to match a printer suitable for the output format %q and the options specified: %#v", output, e.Options)
>>>>>>> axbaretto
}

func IsNoCompatiblePrinterError(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(NoCompatiblePrinterError)
	return ok
}

// PrintFlags composes common printer flag structs
// used across all commands, and provides a method
// of retrieving a known printer based on flag values provided.
type PrintFlags struct {
<<<<<<< HEAD
	JSONYamlPrintFlags   *JSONYamlPrintFlags
	NamePrintFlags       *NamePrintFlags
	TemplatePrinterFlags *KubeTemplatePrintFlags
=======
	JSONYamlPrintFlags *JSONYamlPrintFlags
	NamePrintFlags     *NamePrintFlags
>>>>>>> axbaretto

	TypeSetterPrinter *printers.TypeSetterPrinter

	OutputFormat *string
<<<<<<< HEAD

	// OutputFlagSpecified indicates whether the user specifically requested a certain kind of output.
	// Using this function allows a sophisticated caller to change the flag binding logic if they so desire.
	OutputFlagSpecified func() bool
=======
>>>>>>> axbaretto
}

func (f *PrintFlags) Complete(successTemplate string) error {
	return f.NamePrintFlags.Complete(successTemplate)
}

<<<<<<< HEAD
func (f *PrintFlags) AllowedFormats() []string {
	ret := []string{}
	ret = append(ret, f.JSONYamlPrintFlags.AllowedFormats()...)
	ret = append(ret, f.NamePrintFlags.AllowedFormats()...)
	ret = append(ret, f.TemplatePrinterFlags.AllowedFormats()...)
	return ret
}

=======
>>>>>>> axbaretto
func (f *PrintFlags) ToPrinter() (printers.ResourcePrinter, error) {
	outputFormat := ""
	if f.OutputFormat != nil {
		outputFormat = *f.OutputFormat
	}
<<<<<<< HEAD
	// For backwards compatibility we want to support a --template argument given, even when no --output format is provided.
	// If no explicit output format has been provided via the --output flag, fallback
	// to honoring the --template argument.
	templateFlagSpecified := f.TemplatePrinterFlags != nil &&
		f.TemplatePrinterFlags.TemplateArgument != nil &&
		len(*f.TemplatePrinterFlags.TemplateArgument) > 0
	outputFlagSpecified := f.OutputFlagSpecified != nil && f.OutputFlagSpecified()
	if templateFlagSpecified && !outputFlagSpecified {
		outputFormat = "go-template"
	}
=======
>>>>>>> axbaretto

	if f.JSONYamlPrintFlags != nil {
		if p, err := f.JSONYamlPrintFlags.ToPrinter(outputFormat); !IsNoCompatiblePrinterError(err) {
			return f.TypeSetterPrinter.WrapToPrinter(p, err)
		}
	}

	if f.NamePrintFlags != nil {
		if p, err := f.NamePrintFlags.ToPrinter(outputFormat); !IsNoCompatiblePrinterError(err) {
			return f.TypeSetterPrinter.WrapToPrinter(p, err)
		}
	}

<<<<<<< HEAD
	if f.TemplatePrinterFlags != nil {
		if p, err := f.TemplatePrinterFlags.ToPrinter(outputFormat); !IsNoCompatiblePrinterError(err) {
			return f.TypeSetterPrinter.WrapToPrinter(p, err)
		}
	}

	return nil, NoCompatiblePrinterError{OutputFormat: f.OutputFormat, AllowedFormats: f.AllowedFormats()}
=======
	return nil, NoCompatiblePrinterError{Options: f, OutputFormat: f.OutputFormat}
>>>>>>> axbaretto
}

func (f *PrintFlags) AddFlags(cmd *cobra.Command) {
	f.JSONYamlPrintFlags.AddFlags(cmd)
	f.NamePrintFlags.AddFlags(cmd)
<<<<<<< HEAD
	f.TemplatePrinterFlags.AddFlags(cmd)

	if f.OutputFormat != nil {
		cmd.Flags().StringVarP(f.OutputFormat, "output", "o", *f.OutputFormat, fmt.Sprintf("Output format. One of: %s.", strings.Join(f.AllowedFormats(), "|")))
		if f.OutputFlagSpecified == nil {
			f.OutputFlagSpecified = func() bool {
				return cmd.Flag("output").Changed
			}
		}
=======

	if f.OutputFormat != nil {
		cmd.Flags().StringVarP(f.OutputFormat, "output", "o", *f.OutputFormat, "Output format. One of: json|yaml|wide|name|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=... See custom columns [http://kubernetes.io/docs/user-guide/kubectl-overview/#custom-columns], golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [http://kubernetes.io/docs/user-guide/jsonpath].")
>>>>>>> axbaretto
	}
}

// WithDefaultOutput sets a default output format if one is not provided through a flag value
func (f *PrintFlags) WithDefaultOutput(output string) *PrintFlags {
	f.OutputFormat = &output
	return f
}

// WithTypeSetter sets a wrapper than will surround the returned printer with a printer to type resources
func (f *PrintFlags) WithTypeSetter(scheme *runtime.Scheme) *PrintFlags {
	f.TypeSetterPrinter = printers.NewTypeSetter(scheme)
	return f
}

func NewPrintFlags(operation string) *PrintFlags {
	outputFormat := ""

	return &PrintFlags{
		OutputFormat: &outputFormat,

<<<<<<< HEAD
		JSONYamlPrintFlags:   NewJSONYamlPrintFlags(),
		NamePrintFlags:       NewNamePrintFlags(operation),
		TemplatePrinterFlags: NewKubeTemplatePrintFlags(),
=======
		JSONYamlPrintFlags: NewJSONYamlPrintFlags(),
		NamePrintFlags:     NewNamePrintFlags(operation),
>>>>>>> axbaretto
	}
}
