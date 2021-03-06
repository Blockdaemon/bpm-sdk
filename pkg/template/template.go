// Package template implements functions to render Go templates to files using the node.Node struct as an imnput for the templates.
package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"go.blockdaemon.com/bpm/sdk/pkg/fileutil"
	"go.blockdaemon.com/bpm/sdk/pkg/node"
)

// TemplateData wraps the data send to the rendering engine
type TemplateData struct {
	Node       node.Node
	PluginData map[string]interface{}
}

// ConfigFileRendered renders a template with node confguration and writes it to disk if it doesn't exist yet
//
// In order to allow comma separated lists in the template it defines the template
// function `notLast` which can be used like this:
//
//		{{range $index, $id:= .Config.core.quorum_set_ids -}}
//		"${{ $id }}"{{if notLast $index $.Config.core.quorum_set_ids}},{{end}}
//		{{end -}}
//
func ConfigFileRendered(filepath, templateContent string, templateData TemplateData) error {
	outputFilename := path.Join(templateData.Node.NodeDirectory(), filepath)

	exists, err := fileutil.FileExists(outputFilename)
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("File '%s' already exists, skipping creation\n", outputFilename)
		return nil
	}

	fmt.Printf("Writing file '%s'\n", outputFilename)

	var templateFunctions = template.FuncMap{
		"notLast": func(x int, a []interface{}) bool {
			return x != len(a)-1
		},
	}

	tmpl, err := template.New(outputFilename).Funcs(templateFunctions).Parse(templateContent)
	if err != nil {
		return err
	}

	output := bytes.NewBufferString("")

	err = tmpl.Execute(output, templateData)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputFilename, output.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// ConfigFilesRendered renderes multiple templates to files
func ConfigFilesRendered(filenamesAndTemplates map[string]string, templateData TemplateData) error {
	for filename, template := range filenamesAndTemplates {
		if err := ConfigFileRendered(filename, template, templateData); err != nil {
			return err
		}

	}

	return nil
}

// ConfigFileAbsent deletes a file if it exists
func ConfigFileAbsent(filename string, node node.Node) error {
	filePath := path.Join(node.NodeDirectory(), filename)

	exists, err := fileutil.FileExists(filePath)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Cannot find file '%s', skipping removal\n", filePath)
		return nil
	}

	fmt.Printf("Removing file '%s'\n", filePath)
	return os.Remove(filePath)
}
