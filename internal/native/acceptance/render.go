package nativeacc

import (
	"bytes"
	"fmt"
	"text/template"
)

// tmplData is the variable set available to every config template: the workspace's
// fixed identifiers (random suffixes, locations, subscription) so the shared base
// resources and each resource-under-test agree on names and placement. The fields
// are fixed for the lifetime of a Workspace.
type tmplData struct {
	RandomInteger  int
	RandomString   string
	Location       string
	LocationAlt    string
	SubscriptionID string
}

func render(tpl string, data tmplData) string {
	t, err := template.New("cfg").Parse(tpl)
	if err != nil {
		panic(fmt.Sprintf("nativeacc: invalid template: %v", err))
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("nativeacc: template execution failed: %v", err))
	}
	return buf.String()
}
