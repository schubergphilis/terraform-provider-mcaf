package mcaf

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = New()
	testAccProviders = map[string]*schema.Provider{
		"mcaf": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := New().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = New()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("CISCOASA_API_URL"); v == "" {
		t.Fatal("CISCOASA_API_URL must be set for acceptance tests")
	}
	if v := os.Getenv("CISCOASA_USERNAME"); v == "" {
		t.Fatal("CISCOASA_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("CISCOASA_PASSWORD"); v == "" {
		t.Fatal("CISCOASA_PASSWORD must be set for acceptance tests")
	}
	if v := os.Getenv("CISCOASA_OBJECT_PREFIX"); v == "" {
		t.Fatal("CISCOASA_OBJECT_PREFIX must be set for acceptance tests")
	}
	if v := os.Getenv("CISCOASA_INTERFACE_NAME"); v == "" {
		t.Fatal("CISCOASA_INTERFACE_NAME must be set for acceptance tests")
	}
}

// Defines a prefix used to mark resources created by the acceptance tests.
var CISCOASA_OBJECT_PREFIX = os.Getenv("CISCOASA_OBJECT_PREFIX")

// The interface name (nameif) of an existing interface.
var CISCOASA_INTERFACE_NAME = os.Getenv("CISCOASA_INTERFACE_NAME")
