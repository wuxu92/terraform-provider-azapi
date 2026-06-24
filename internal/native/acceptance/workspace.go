package nativeacc

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

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
// azapi provider, scoped to one top-level (Ordered) Ginkgo Describe.
//
// Resources are organised into Scopes — one per Ordered container. The root scope
// holds workspace-wide base resources; nested scopes (ws.Scope() / Scope.Scope())
// own the resources of nested containers. A scope is provisioned in its container's
// BeforeAll/It hooks and torn down in its AfterAll, destroying ONLY its own
// resources, so a nested container reuses the base resources of the containers that
// enclose it (referenced by Terraform address).
//
// Terraform drives the provider via reattach (TF_REATTACH_PROVIDERS): the provider
// runs in this test process — the SAME build under test — so no provider binary is
// installed and the schema exercised is exactly the generated one.
type Workspace struct {
	root *Scope

	dir      string
	tf       *tfexec.Terraform
	reattach tfexec.ReattachInfo
	cancel   context.CancelFunc
	closeCh  <-chan struct{}

	td      acceptance.TestData
	client  *clients.Client
	started bool
}

type trackedResource struct {
	armType    string
	apiVersion string
	id         string
}

// NewWorkspace returns an empty workspace: one Terraform working directory plus an
// in-process azapi provider. Declare resources per scope — ws.Resource for the root
// scope, ws.Scope() for nested containers — and apply them in BeforeAll/It hooks.
func NewWorkspace() *Workspace {
	ws := &Workspace{}
	ws.root = &Scope{ws: ws}
	return ws
}

// Resource declares a resource-under-test in the workspace's root scope; it is torn
// down with the whole workspace in Destroy.
func (w *Workspace) Resource(tfType, label string) *Resource {
	return w.root.Resource(tfType, label)
}

// ResourceFor declares a resource-under-test in the root scope from a config builder
// (see Scope.ResourceFor and ResourceConfig).
func (w *Workspace) ResourceFor(c ResourceConfig) *Resource {
	return w.root.ResourceFor(c)
}

// Scope returns a child of the root scope. Wire its Teardown to a nested container's
// AfterAll so the container's own resources are destroyed while the base survives.
func (w *Workspace) Scope() *Scope {
	return w.root.Scope()
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
	// Silence the in-process provider's per-RPC trace logging by default.
	quietProviderLogs()

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

	gomega.Expect(tf.Init(context.Background(), tfexec.Reattach(w.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform init")
	// From here Destroy must run even though no resources exist yet; scopes apply
	// their resources in later BeforeAll / It hooks.
	w.started = true
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
	for _, tr := range w.root.tracked {
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

// providerLogEnvVars are the tfsdklog/tflog streams the in-process azapi provider emits
// to the test process's stderr: the per-RPC sdk.proto traces ("Received request"), the
// SDK and framework streams, and the root provider stream (tflog). terraform-plugin-go
// reads each stream's level straight from its own env var — TF_LOG is not consulted for
// an in-process reattach provider — and an unset var resolves to a level that prints at
// trace/info, hence the flood.
var providerLogEnvVars = []string{
	"TF_LOG_SDK_PROTO",
	"TF_LOG_SDK",
	"TF_LOG_SDK_FRAMEWORK",
	// The root provider stream (tflog @module=provider — e.g. the timeouts library's
	// "read timeout configuration not found, using provided default") is gated by a
	// NAME-SUFFIXED var, not bare TF_LOG_PROVIDER: tf6server composes
	// ToUpper("TF_LOG_PROVIDER" + "_" + ProviderLoggerName(addr)). plugin.DebugServe
	// defaults addr to "provider", so the real gate is TF_LOG_PROVIDER_PROVIDER. Bare
	// TF_LOG_PROVIDER is kept too — it only matters for a subprocess provider.
	"TF_LOG_PROVIDER",
	"TF_LOG_PROVIDER_PROVIDER",
}

// quietProviderLogs makes a normal acceptance run quiet by silencing the two log
// channels the in-process azapi provider writes to the test process's stderr:
//
//   - the tfsdklog streams above (sdk.proto RPC traces, SDK, framework, provider),
//     via their per-stream env vars; and
//   - the stdlib log package, where the azcore event listener and the live-traffic
//     policy print every Azure HTTP request/response (internal/clients). A go-plugin
//     subprocess captures that and gates it on TF_LOG_PROVIDER, but in reattach mode
//     it lands straight on stderr, uncaptured.
//
// It honors an explicit per-stream level if one is already set, and treats TF_LOG as
// the opt-in master switch these in-process subsystems otherwise ignore: run with
// TF_LOG=trace (or debug/info/...) to restore full provider/RPC/HTTP logging.
func quietProviderLogs() {
	level := "off"
	quiet := true
	if v := os.Getenv("TF_LOG"); v != "" {
		level = v
		quiet = false
	}
	for _, v := range providerLogEnvVars {
		if os.Getenv(v) == "" {
			_ = os.Setenv(v, level)
		}
	}
	// Redirect the stdlib log default logger (azcore listener + live-traffic policy)
	// to the void by default; send it back to stderr when TF_LOG opts in.
	if quiet {
		log.SetOutput(io.Discard)
	} else {
		log.SetOutput(os.Stderr)
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
