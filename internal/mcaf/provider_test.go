package mcaf

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// Provider name for single configuration testing
	ProviderName = "mcaf"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	// Always allocate a new provider instance each invocation, otherwise gRPC
	// ProviderConfigure() can overwrite configuration during concurrent testing.
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderName: func() (*schema.Provider, error) { return New(), nil }, //nolint:unparam
	}
}

func TestProvider(t *testing.T) {
	if err := New().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = New()
}

func testAccAwsPreCheck(t *testing.T) {
	if v := os.Getenv("AWS_ACCESS_KEY_ID"); v == "" {
		t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
	}
	if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v == "" {
		t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
	}
}

// testAccAwsClient configures and returns a fully initialized AWSClient.
func testAccAwsClient() (*AWSClient, error) {
	region := "us-east-1"
	if v := os.Getenv("AWS_DEFAULT_REGION"); v != "" {
		region = v
	}

	aws := map[string]interface{}{
		"access_key":                  "",
		"secret_key":                  "",
		"profile":                     "",
		"region":                      region,
		"max_retries":                 3,
		"shared_credentials_file":     "",
		"skip_credentials_validation": false,
		"skip_metadata_api_check":     false,
		"skip_requesting_account_id":  false,
		"token":                       "",
	}

	client, err := awsClient(aws)
	if err != nil {
		return nil, err
	}

	return client, nil
}
