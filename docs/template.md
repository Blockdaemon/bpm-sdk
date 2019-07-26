# template
--
    import "gitlab.com/Blockdaemon/bpm-sdk/pkg/template"

Package template implements functions to render Go templates to files using the
node.Node struct as an imnput for the templates.

## Usage

#### func  ConfigFileAbsent

```go
func ConfigFileAbsent(filename string, node node.Node) error
```
ConfigFileAbsent deletes a file if it exists

#### func  ConfigFileRendered

```go
func ConfigFileRendered(filename, templateContent string, node node.Node) error
```
ConfigFileRendered renders a template with node confguration and writes it to
disk if it doesn't exist yet

In order to allow comma separated lists in the template it defines the template
function `notLast` which can be used like this:

    {{range $index, $id:= .Config.core.quorum_set_ids -}}
    "${{ $id }}"{{if notLast $index $.Config.core.quorum_set_ids}},{{end}}
    {{end -}}
