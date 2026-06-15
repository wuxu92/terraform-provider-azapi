// Package azapinacc is a BDD-style acceptance-test framework for azapin static
// resources. It wraps terraform-plugin-testing with Ginkgo so that many
// scenarios for a resource live in one suite and share a single config
// template — only the resource body differs per scenario.
//
// Usage (one suite per resource):
//
//	func TestStorageAccount(t *testing.T) {
//	    RegisterFailHandler(Fail)
//	    RunSpecs(t, "azapi_storage_account")
//	}
//
//	var _ = Describe("azapi_storage_account", func() {
//	    spec := azapinacc.NewSpec("azapi_storage_account",
//	        "Microsoft.Storage/storageAccounts", "2025-01-01")
//	    It("creates, updates and imports", func() {
//	        spec.Run(
//	            azapinacc.Body(`location = "{{.Location}}" ... `).
//	                Check(azapinacc.Exists(), azapinacc.Key("kind").HasValue("StorageV2")),
//	            azapinacc.ImportStep(),
//	        )
//	    })
//	})
package azapinacc

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/Azure/terraform-provider-azapi/internal/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	ginkgo "github.com/onsi/ginkgo/v2"
)

// Spec describes one azapin resource under test: its Terraform type, the ARM
// type + API version it targets, and the shared parent/config template.
type Spec struct {
	tfType     string
	armType    string
	apiVersion string
	label      string
	parentTpl  string
	nameFn     func(acceptance.TestData) string
}

// NewSpec builds a Spec for a generated resource. The defaults use an
// azapi_resource resource group as the parent and a lowercase-alphanumeric name
// (safe for storage-style naming). Override with WithParent / WithName.
func NewSpec(tfType, armType, apiVersion string) *Spec {
	return &Spec{
		tfType:     tfType,
		armType:    armType,
		apiVersion: apiVersion,
		label:      "test",
		parentTpl:  defaultParentTemplate,
		nameFn:     defaultName,
	}
}

// WithParent overrides the parent template. The template must declare a resource
// labelled "parent" whose `.id` is referenced as the resource's parent_id.
func (s *Spec) WithParent(tpl string) *Spec { s.parentTpl = tpl; return s }

// WithName overrides the resource-name generator.
func (s *Spec) WithName(fn func(acceptance.TestData) string) *Spec { s.nameFn = fn; return s }

// Step is one scenario step: a resource body fragment plus checks, an import
// verification, or an expected-error apply.
type Step struct {
	body        string
	checks      []Check
	expectError string
	importStep  bool
	ignore      []string
}

// Body starts a config step from a resource-body fragment. The fragment is HCL
// for the attributes inside the resource block and may use template variables
// such as {{.Location}}, {{.RandomInteger}}, {{.RandomString}}.
func Body(b string) *Step { return &Step{body: b} }

// Check attaches assertions to a config step.
func (st *Step) Check(c ...Check) *Step { st.checks = append(st.checks, c...); return st }

// ExpectError marks a config step as expected to fail apply with a matching error.
func (st *Step) ExpectError(re string) *Step { st.expectError = re; return st }

// ImportStep returns a step that imports the resource and verifies state parity.
// ignore lists attribute paths to skip during verification (e.g. write-only).
func ImportStep(ignore ...string) *Step { return &Step{importStep: true, ignore: ignore} }

// Run assembles and executes the full terraform-plugin-testing lifecycle for the
// given steps. It skips when TF_ACC is unset.
func (s *Spec) Run(steps ...*Step) {
	if os.Getenv("TF_ACC") == "" || os.Getenv("ARM_SUBSCRIPTION_ID") == "" {
		ginkgo.Skip("acceptance tests skipped: set TF_ACC=1 and ARM_SUBSCRIPTION_ID (+ credentials)")
	}

	td := s.buildData()
	name := s.nameFn(td)
	tr := newExists(s.armType, s.apiVersion)
	cc := &checkCtx{resourceName: td.ResourceName, tr: tr}

	tfSteps := make([]resource.TestStep, 0, len(steps))
	for _, st := range steps {
		if st.importStep {
			tfSteps = append(tfSteps, resource.TestStep{
				ResourceName:            td.ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: st.ignore,
			})
			continue
		}
		ts := resource.TestStep{Config: s.renderConfig(td, name, st.body)}
		if len(st.checks) > 0 {
			fns := make([]resource.TestCheckFunc, 0, len(st.checks))
			for _, c := range st.checks {
				fns = append(fns, c(cc))
			}
			ts.Check = resource.ComposeTestCheckFunc(fns...)
		}
		if st.expectError != "" {
			ts.ExpectError = regexp.MustCompile(st.expectError)
		}
		tfSteps = append(tfSteps, ts)
	}

	testCase := resource.TestCase{
		PreCheck:                 func() { preCheck() },
		ProtoV6ProviderFactories: td.Providers(),
		CheckDestroy: func(state *terraform.State) error {
			client, err := acceptance.BuildTestClient()
			if err != nil {
				return fmt.Errorf("building test client: %w", err)
			}
			return acceptance.CheckDestroyedFunc(client, tr, td.ResourceType, td.ResourceName)(state)
		},
		Steps: tfSteps,
	}

	resource.Test(ginkgo.GinkgoT(), testCase)
}

// ---------------------------------------------------------------------------
// config rendering
// ---------------------------------------------------------------------------

type tmplData struct {
	RandomInteger  int
	RandomString   string
	Location       string
	LocationAlt    string
	SubscriptionID string
	Name           string
	Type           string
	Label          string
	ParentRef      string
	Body           string
}

const defaultParentTemplate = `
resource "azapi_resource" "parent" {
  type      = "Microsoft.Resources/resourceGroups@2021-04-01"
  name      = "acctest-rg-{{.RandomInteger}}"
  parent_id = "/subscriptions/{{.SubscriptionID}}"
  location  = "{{.Location}}"
}
`

const resourceBlockTemplate = `
resource "{{.Type}}" "{{.Label}}" {
  name      = "{{.Name}}"
  parent_id = {{.ParentRef}}
{{.Body}}
}
`

func (s *Spec) renderConfig(td acceptance.TestData, name, body string) string {
	data := tmplData{
		RandomInteger:  td.RandomInteger,
		RandomString:   td.RandomString,
		Location:       td.LocationPrimary,
		LocationAlt:    td.LocationSecondary,
		SubscriptionID: os.Getenv("ARM_SUBSCRIPTION_ID"),
		Name:           name,
		Type:           s.tfType,
		Label:          s.label,
		ParentRef:      "azapi_resource.parent.id",
	}
	// Render the body fragment first so it may use the same template variables,
	// then inject the result into the resource block (avoids double-render).
	data.Body = indent(render(body, data), "  ")

	var sb strings.Builder
	sb.WriteString(render(s.parentTpl, data))
	sb.WriteString(render(resourceBlockTemplate, data))
	return sb.String()
}

func render(tpl string, data tmplData) string {
	t, err := template.New("cfg").Parse(tpl)
	if err != nil {
		panic(fmt.Sprintf("azapinacc: invalid template: %v", err))
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("azapinacc: template execution failed: %v", err))
	}
	return buf.String()
}

func indent(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

func defaultName(td acceptance.TestData) string {
	name := strings.ToLower(fmt.Sprintf("acc%s%d", td.RandomString, td.RandomInteger))
	if len(name) > 24 {
		name = name[:24]
	}
	return name
}

// buildData constructs a minimal TestData without azapi's reader-client / triple-
// location requirements. Only the fields the framework actually uses are set.
func (s *Spec) buildData() acceptance.TestData {
	loc := os.Getenv("ARM_TEST_LOCATION")
	if loc == "" {
		loc = "westeurope"
	}
	locAlt := os.Getenv("ARM_TEST_LOCATION_ALT")
	if locAlt == "" {
		locAlt = "eastus"
	}
	return acceptance.TestData{
		RandomInteger:     acceptance.RandTimeInt(),
		RandomString:      acctest.RandStringFromCharSet(5, "abcdefghijklmnopqrstuvwxyz0123456789"),
		ResourceType:      s.tfType,
		ResourceName:      s.tfType + "." + s.label,
		LocationPrimary:   location.Normalize(loc),
		LocationSecondary: location.Normalize(locAlt),
	}
}

func preCheck() {
	for _, v := range []string{"ARM_SUBSCRIPTION_ID"} {
		if os.Getenv(v) == "" {
			ginkgo.Skip(v + " must be set for acceptance tests")
		}
	}
}
