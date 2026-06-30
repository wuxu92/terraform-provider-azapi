package web

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/resources"
)

// WebSiteCfg carries the Terraform address metadata and dependencies for
// azapi_web_site acceptance-test scenarios. Construct it with NewWebSiteCfg,
// then wrap it in a scenario type when applying.
type WebSiteCfg struct {
	generated.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
	serverFarm    WebServerFarmCfg
}

// NewWebSiteCfg builds a Web site config depending on the parent resource group
// and the native App Service plan config. The label is optional.
func NewWebSiteCfg(resourceGroup resources.ResourceGroupCfg, serverFarm WebServerFarmCfg, label ...string) WebSiteCfg {
	return WebSiteCfg{
		ResourceConfigBase: generated.NewResourceConfigBase(WebSite.Name, label...),
		resourceGroup:      resourceGroup,
		serverFarm:         serverFarm,
	}
}

// WebSiteCfg_Basic is a minimal Windows App Service site bound to the native
// server farm dependency.
type WebSiteCfg_Basic WebSiteCfg

func (r WebSiteCfg_Basic) Config() string {
	return WebSiteCfg(r).config(`
  properties = {
    server_farm_id = %s
  }`)
}

// WebSiteCfg_Complete exercises a broad Windows App Service configuration that
// only depends on the server farm and resource group created by this package.
// Subresource-backed settings such as app settings, auth, backup, storage mounts,
// publishing policies, and connection strings stay out of this scenario because
// Microsoft.Web/sites does not round-trip them through the sites/config/web read.
type WebSiteCfg_Complete WebSiteCfg

func (r WebSiteCfg_Complete) Config() string {
	return WebSiteCfg(r).config(`
  identity = {
    type = "SystemAssigned"
  }

  tags = {
    environment = "acctest"
    scenario    = "complete"
  }

  properties = {
    server_farm_id                       = %s
    client_affinity_enabled              = true
    client_affinity_partitioning_enabled = false
    client_affinity_proxy_enabled        = false
    client_cert_enabled                  = true
    client_cert_exclusion_paths          = "/health"
    client_cert_mode                     = "Optional"
    daily_memory_time_quota              = 0
    enabled                              = true
    end_to_end_encryption_enabled        = false
    host_names_disabled                  = false
    https_only                           = true
    hyper_v                              = false
    is_xenon                             = false
    public_network_access                = "Enabled"
    reserved                             = false
    scm_site_also_stopped                = false
    ssh_enabled                          = false

    outbound_vnet_routing = {
      all_traffic            = false
      application_traffic    = false
      backup_restore_traffic = false
      content_share_traffic  = false
      image_pull_traffic     = false
    }

    site_config = {
      acr_use_managed_identity_creds = false
      always_on                      = true
      app_command_line               = "echo azapi"
      auto_heal_enabled              = true
      detailed_error_logging_enabled = true
      ftps_state                     = "FtpsOnly"
      health_check_path              = "/health"
      http20_enabled                 = true
      http20_proxy_flag              = 0
      http_logging_enabled           = true
      load_balancing                 = "LeastRequests"
      local_my_sql_enabled           = false
      logs_directory_size_limit      = 35
      managed_pipeline_mode          = "Integrated"
      min_tls_cipher_suite           = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
      min_tls_version                = "1.2"
      net_framework_version          = "v4.0"
      number_of_workers              = 1
      public_network_access          = "Enabled"
      remote_debugging_enabled       = true
      remote_debugging_version       = "VS2022"
      request_tracing_enabled        = true
      scm_min_tls_version            = "1.2"
      scm_type                       = "None"
      use32_bit_worker_process       = false
      vnet_route_all_enabled         = false
      web_sockets_enabled            = true
      website_time_zone              = "UTC"

      api_definition = {
        url = "https://example.com/openapi.json"
      }

      auto_heal_rules = {
        actions = {
          action_type                = "Recycle"
          min_process_execution_time = "00:05:00"
        }
        triggers = {
          status_codes = [{
            count         = 10
            status        = 500
            time_interval = "00:01:00"
          }]
        }
      }

      cors = {
        allowed_origins = [
          "https://www.contoso.com",
          "https://contoso.com",
        ]
        support_credentials = true
      }

      default_documents = [
        "first.html",
        "second.jsp",
        "third.aspx",
        "hostingstart.html",
      ]

      handler_mappings = [{
        extension        = "htm"
        script_processor = "C:\\Program Files (x86)\\Common Files\\Microsoft Shared\\Phone Tools\\11.0\\WebResources\\Microsoft.Web.Deployment\\3.6.0\\msdeploy.axd"
      }]

      ip_security_restrictions = [{
        action      = "Deny"
        description = "Deny documentation range"
        ip_address  = "192.0.2.0/24"
        name        = "deny-doc-range"
        priority    = 100
        tag         = "Default"
      }]
      ip_security_restrictions_default_action = "Allow"

      scm_ip_security_restrictions = [{
        action      = "Deny"
        description = "Deny documentation range"
        ip_address  = "198.51.100.0/24"
        name        = "deny-scm-doc-range"
        priority    = 100
        tag         = "Default"
      }]
      scm_ip_security_restrictions_default_action = "Allow"
      scm_ip_security_restrictions_use_main       = false

      virtual_applications = [{
        virtual_path    = "/"
        physical_path   = "site\\wwwroot"
        preload_enabled = true
        virtual_directories = [{
          virtual_path  = "/assets"
          physical_path = "site\\wwwroot\\assets"
        }]
      }]
    }
  }`)
}

func (r WebSiteCfg) config(bodyFmt string) string {
	return fmt.Sprintf(`
resource %q %q {
  name              = "acctest-web-{{.RandomString}}"
  resource_group_id = %s
  location          = "{{.Location}}"
  kind              = "app"%s
}
`, r.ResourceType(), r.ResourceLabel(), r.resourceGroup.IDRef(), fmt.Sprintf(bodyFmt, r.serverFarm.IDRef()))
}
