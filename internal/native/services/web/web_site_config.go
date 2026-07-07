package web

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// WebSiteCfg carries the Terraform address metadata and dependencies for
// azapi_web_site acceptance-test scenarios. Construct it with NewWebSiteCfg,
// then wrap it in a scenario type when applying.
type WebSiteCfg struct {
	services.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
	serverFarm    WebServerFarmCfg
}

// NewWebSiteCfg builds a Web site config depending on the parent resource group
// and the native App Service plan config. The label is optional.
func NewWebSiteCfg(resourceGroup resources.ResourceGroupCfg, serverFarm WebServerFarmCfg, label ...string) WebSiteCfg {
	return WebSiteCfg{
		ResourceConfigBase: services.NewResourceConfigBase(WebSite.Name, label...),
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

// WebSiteCfg_Complete_update mutates every in-place-updatable property set by
// WebSiteCfg_Complete to a different valid value. Applying Complete then
// Complete_update proves those properties round-trip through an in-place update
// (Update -> config/web read -> empty plan) and that Azure accepts the new values.
// Interdependent toggles are held stable on purpose (e.g. client_cert_enabled and
// http_logging_enabled stay true so client_cert_mode and logs_directory_size_limit
// keep round-tripping; auto_heal_enabled stays true so its rules survive the read),
// while the dependent value under each is what changes.
type WebSiteCfg_Complete_update WebSiteCfg

func (r WebSiteCfg_Complete_update) Config() string {
	return WebSiteCfg(r).config(`
  identity = {
    type = "SystemAssigned"
  }

  tags = {
    environment = "acctest-update"
    scenario    = "complete-update"
  }

  properties = {
    server_farm_id                       = %s
    client_affinity_enabled              = false
    client_affinity_partitioning_enabled = false
    client_affinity_proxy_enabled        = false
    client_cert_enabled                  = true
    client_cert_exclusion_paths          = "/healthz"
    client_cert_mode                     = "Required"
    daily_memory_time_quota              = 0
    enabled                              = true
    end_to_end_encryption_enabled        = false
    host_names_disabled                  = false
    https_only                           = false
    hyper_v                              = false
    is_xenon                             = false
    public_network_access                = "Disabled"
    reserved                             = false
    scm_site_also_stopped                = false // held: ARM only honors true while the app is stopped, so a running site always reads back false (AzureRM omits this property entirely)
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
      always_on                      = false
      app_command_line               = "echo updated"
      auto_heal_enabled              = true
      detailed_error_logging_enabled = false
      ftps_state                     = "AllAllowed"
      health_check_path              = "/healthz"
      http20_enabled                 = false
      http20_proxy_flag              = 0
      http_logging_enabled           = true
      load_balancing                 = "WeightedRoundRobin"
      local_my_sql_enabled           = false
      logs_directory_size_limit      = 50
      managed_pipeline_mode          = "Classic"
      min_tls_cipher_suite           = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
      min_tls_version                = "1.2"
      net_framework_version          = "v6.0"
      number_of_workers              = 1
      public_network_access          = "Disabled"
      remote_debugging_enabled       = true
      remote_debugging_version       = "VS2022"
      request_tracing_enabled        = false
      scm_min_tls_version            = "1.2"
      scm_type                       = "None"
      use32_bit_worker_process       = true
      vnet_route_all_enabled         = false
      web_sockets_enabled            = false
      website_time_zone              = "UTC" // held: WEBSITE_TIME_ZONE app setting takes precedence over this config field (per ARM), so a non-default value reads back UTC (AzureRM omits this property entirely)

      api_definition = {
        url = "https://example.com/updated-openapi.json"
      }

      auto_heal_rules = {
        actions = {
          action_type                = "Recycle"
          min_process_execution_time = "00:10:00"
        }
        triggers = {
          status_codes = [{
            count         = 5
            status        = 501
            time_interval = "00:05:00"
          }]
        }
      }

      cors = {
        allowed_origins = [
          "https://updated.contoso.com",
        ]
        support_credentials = false
      }

      default_documents = [
        "updated.html",
        "index.html",
        "hostingstart.html",
      ]

      handler_mappings = [{
        extension        = "html"
        script_processor = "C:\\Program Files (x86)\\Common Files\\Microsoft Shared\\Phone Tools\\11.0\\WebResources\\Microsoft.Web.Deployment\\3.6.0\\msdeploy.axd"
      }]

      ip_security_restrictions = [{
        action      = "Deny"
        description = "Deny updated documentation range"
        ip_address  = "203.0.113.0/24"
        name        = "deny-updated-range"
        priority    = 200
        tag         = "Default"
      }]
      ip_security_restrictions_default_action = "Deny"

      scm_ip_security_restrictions = [{
        action      = "Deny"
        description = "Deny updated documentation range"
        ip_address  = "203.0.113.128/25"
        name        = "deny-updated-scm-range"
        priority    = 200
        tag         = "Default"
      }]
      scm_ip_security_restrictions_default_action = "Deny"
      scm_ip_security_restrictions_use_main       = false

      virtual_applications = [{
        virtual_path    = "/"
        physical_path   = "site\\wwwroot"
        preload_enabled = false
        virtual_directories = [{
          virtual_path  = "/static"
          physical_path = "site\\wwwroot\\static"
        }]
      }]
    }
  }`)
}

// WebSiteCfg_Identity assigns a user-assigned identity to the site, exercising the
// root-level identity block (type=UserAssigned) whose user_assigned_identities map is
// keyed by a native azapi_user_assigned_identity's ARM resource ID. Identity is that
// identity's IDRef (e.g. "azapi_user_assigned_identity.test.id").
type WebSiteCfg_Identity struct {
	WebSiteCfg
	Identity string
}

func (r WebSiteCfg_Identity) Config() string {
	// Substitute the identity ref here; %%s stays literal so the shared config() helper
	// fills server_farm_id in its own Sprintf pass.
	body := fmt.Sprintf(`
  identity = {
    type = "UserAssigned"
    user_assigned_identities = {
      (%s) = {}
    }
  }

  properties = {
    server_farm_id = %%s
  }`, r.Identity)
	return r.WebSiteCfg.config(body)
}

func (r WebSiteCfg) config(bodyFmt string) string {
	return r.RenderConfig(services.ConfigEnvelope{
		Name:       "acctest-web-{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Kind:       "app",
		Body:       fmt.Sprintf(bodyFmt, r.serverFarm.IDRef()),
	})
}
