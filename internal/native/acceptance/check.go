package azapinacc

import (
	"github.com/Azure/terraform-provider-azapi/internal/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/acceptance/check"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Check is a scenario assertion bound to the resource under test.
type Check func(c *checkCtx) resource.TestCheckFunc

type checkCtx struct {
	resourceName string
	tr           acceptance.TestResource
}

// Exists asserts the resource exists in Azure (GET succeeds).
func Exists() Check {
	return func(c *checkCtx) resource.TestCheckFunc {
		return check.That(c.resourceName).ExistsInAzure(c.tr)
	}
}

// KeyAssert builds attribute-level assertions for a Terraform state path,
// e.g. Key("properties.access_tier").HasValue("Hot").
type KeyAssert struct{ key string }

// Key starts an attribute assertion for the given Terraform state path.
func Key(key string) KeyAssert { return KeyAssert{key: key} }

// HasValue asserts the attribute equals value.
func (k KeyAssert) HasValue(value string) Check {
	return func(c *checkCtx) resource.TestCheckFunc {
		return check.That(c.resourceName).Key(k.key).HasValue(value)
	}
}

// IsSet asserts the attribute is set (non-empty).
func (k KeyAssert) IsSet() Check {
	return func(c *checkCtx) resource.TestCheckFunc {
		return check.That(c.resourceName).Key(k.key).IsSet()
	}
}

// Exists asserts the attribute is present in state.
func (k KeyAssert) Exists() Check {
	return func(c *checkCtx) resource.TestCheckFunc {
		return check.That(c.resourceName).Key(k.key).Exists()
	}
}
