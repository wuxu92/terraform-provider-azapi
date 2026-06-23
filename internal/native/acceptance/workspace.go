package nativeacc

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/provider"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

// Workspace is a single shared Terraform working directory plus an in-process
// azapi provider, scoped to one (Ordered) Ginkgo Describe.
//
// The base resources are applied once in Start (BeforeAll); each It applies its
// own resource-under-test into the SAME working directory and asserts; everything
// is destroyed once in Destroy (AfterAll). This mirrors the azurerm-style
// "base + scenario steps" model while keeping every It focused on one resource,
// and avoids re-creating the resource group / storage account per scenario.
//
// Terraform drives the provider via reattach (TF_REATTACH_PROVIDERS): the provider
// runs in this test process — the SAME build under test — so no provider binary is
// installed and the schema exercised is exactly the generated one.
type Workspace struct {
	base string

	dir      string
	tf       *tfexec.Terraform
	reattach tfexec.ReattachInfo
	cancel   context.CancelFunc
	closeCh  <-chan struct{}

	td      acceptance.TestData
	client  *clients.Client
	tracked []trackedResource
	started bool
}

type trackedResource struct {
	armType    string
	apiVersion string
	id         string
}

// NewWorkspace returns a workspace whose shared base resources are described by
// baseHCL (template-rendered with the standard variables). Pass "" when the
// resource-under-test has no provisioned parent (e.g. a subscription-scoped
// resource group).
func NewWorkspace(baseHCL string) *Workspace {
	return &Workspace{base: baseHCL}
}

// providerConfig configures the in-process azapi provider; credentials are read
// from the standard ARM_* environment variables.
const providerConfig = `provider "azapi" {
}
`

// Start provisions the shared base resources. Intended for BeforeAll. It skips the
// whole Describe when the acceptance preconditions are unmet.
func (w *Workspace) Start() {
	skipIfNotAcc()
	execPath := terraformExecPath()
	// Suppress Terraform's non-blocking checkpoint.hashicorp.com telemetry call so
	// init/apply do not reach out to the network in CI / air-gapped runs.
	_ = os.Setenv("CHECKPOINT_DISABLE", "1")

	w.td = buildData()

	client, err := acceptance.BuildTestClient()
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "building test client")
	w.client = client

	dir, err := os.MkdirTemp("", "nativeacc-")
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "creating work dir")
	w.dir = dir

	// Serve the provider in-process for the whole Describe. Terraform reattaches to
	// it instead of installing a provider binary; the server lives until Destroy.
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	cfg, closeCh, err := plugin.DebugServe(ctx, &plugin.ServeOpts{
		GRPCProviderV6Func:  providerserver.NewProtocol6(provider.AzureProvider()),
		Logger:              hclog.NewNullLogger(),
		NoLogOutputOverride: true,
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "serving provider")
	w.closeCh = closeCh
	w.reattach = reattachInfo(cfg)

	tf, err := tfexec.NewTerraform(dir, execPath)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "constructing terraform")
	w.tf = tf

	w.writeFile("provider.tf", providerConfig)
	if strings.TrimSpace(w.base) != "" {
		w.writeFile("base.tf", w.render(w.base))
	}

	gomega.Expect(tf.Init(context.Background(), tfexec.Reattach(w.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform init")
	// From here Destroy must run even if the base apply fails partway.
	w.started = true
	gomega.Expect(tf.Apply(context.Background(), tfexec.Reattach(w.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform apply (base)")
}

// Destroy tears down the whole workspace. Intended for AfterAll. It always stops
// the provider server and removes the working directory, even if destroy fails.
func (w *Workspace) Destroy() {
	defer func() {
		if w.cancel != nil {
			w.cancel()
			if w.closeCh != nil {
				<-w.closeCh
			}
		}
		if w.dir != "" {
			_ = os.RemoveAll(w.dir)
		}
	}()

	if !w.started || w.tf == nil {
		return
	}

	gomega.Expect(w.tf.Destroy(context.Background(), tfexec.Reattach(w.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform destroy")

	// Confirm every applied resource is really gone from Azure.
	for _, tr := range w.tracked {
		if tr.id == "" {
			continue
		}
		ok, err := azureExists(w.client, tr.armType, tr.apiVersion, tr.id)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "post-destroy GET %s", tr.id)
		gomega.Expect(ok).To(gomega.BeFalse(), "%s still exists after destroy", tr.id)
	}
}

func (w *Workspace) writeFile(name, content string) {
	gomega.Expect(os.WriteFile(filepath.Join(w.dir, name), []byte(content), 0o600)).
		NotTo(gomega.HaveOccurred(), "writing %s", name)
}

func (w *Workspace) render(tpl string) string { return render(tpl, w.tmplData()) }

func (w *Workspace) tmplData() tmplData {
	return tmplData{
		RandomInteger:  w.td.RandomInteger,
		RandomString:   w.td.RandomString,
		Location:       w.td.LocationPrimary,
		LocationAlt:    w.td.LocationSecondary,
		SubscriptionID: os.Getenv("ARM_SUBSCRIPTION_ID"),
	}
}

func (w *Workspace) track(armType, apiVersion, id string) {
	w.tracked = append(w.tracked, trackedResource{armType: armType, apiVersion: apiVersion, id: id})
}

// reattachInfo builds the TF_REATTACH_PROVIDERS payload for the served provider,
// registered under every namespace Terraform might resolve "azapi" to. Reattach
// short-circuits provider installation regardless of the configured source.
func reattachInfo(cfg plugin.ReattachConfig) tfexec.ReattachInfo {
	rc := tfexec.ReattachConfig{
		Protocol:        cfg.Protocol,
		ProtocolVersion: cfg.ProtocolVersion,
		Pid:             cfg.Pid,
		Test:            cfg.Test,
		Addr: tfexec.ReattachConfigAddr{
			Network: cfg.Addr.Network,
			String:  cfg.Addr.String,
		},
	}
	info := tfexec.ReattachInfo{}
	for _, addr := range []string{
		"registry.terraform.io/hashicorp/azapi",
		"registry.terraform.io/-/azapi",
		"registry.terraform.io/Azure/azapi",
	} {
		info[addr] = rc
	}
	return info
}

func terraformExecPath() string {
	if p := os.Getenv("TF_ACC_TERRAFORM_PATH"); p != "" {
		return p
	}
	if p, err := exec.LookPath("terraform"); err == nil {
		return p
	}
	ginkgo.Skip("terraform binary not found; set TF_ACC_TERRAFORM_PATH or install terraform")
	return ""
}

func skipIfNotAcc() {
	if os.Getenv("TF_ACC") == "" || os.Getenv("ARM_SUBSCRIPTION_ID") == "" {
		ginkgo.Skip("acceptance tests skipped: set TF_ACC=1 and ARM_SUBSCRIPTION_ID (+ credentials)")
	}
}

func buildData() acceptance.TestData {
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
		LocationPrimary:   location.Normalize(loc),
		LocationSecondary: location.Normalize(locAlt),
	}
}
