// Package mcaf implements a custom MCAF Terraform provider.
package mcaf

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"aws": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     awsProviderSchema(),
			},

			"o365": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     o365ProviderSchema(),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"mcaf_aws_account": resourceAWSAccount(),
			"mcaf_o365_alias":  resourceO365Alias(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	mcaf := &Client{}

	if aws, ok := d.GetOk("aws"); ok {
		client, err := awsClient(aws.([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		mcaf.AWSClient = client
	}

	if o365, ok := d.GetOk("o365"); ok {
		client, err := o365Client(o365.([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		mcaf.O365 = client
	}

	return mcaf, nil
}

func checkProvider(p string, f func(*schema.ResourceData, interface{}) error) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		mcaf := meta.(*Client)

		switch p {
		case "aws":
			if mcaf.AWSClient == nil {
				return fmt.Errorf("Missing AWS provider configuration")
			}
		case "o365":
			if mcaf.O365 == nil {
				return fmt.Errorf("Missing O365 provider configuration")
			}
		default:
			return fmt.Errorf("Trying to use unknown provider: %s", p)
		}

		return f(d, meta)
	}
}

func awsProviderSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"profile": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"shared_credentials_file": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"token": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"region": {
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.EnvDefaultFunc("AWS_DEFAULT_REGION", nil),
				InputDefault: "us-east-1",
			},

			"max_retries": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  25,
			},

			"skip_credentials_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"skip_region_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"skip_requesting_account_id": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"skip_metadata_api_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func o365ProviderSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"acl_guid": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("O365_ACL_GUID", nil),
			},

			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("O365_EXOAPI_ENDPOINT", nil),
			},

			"secret_code": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("O365_SECRET_CODE", nil),
			},
		},
	}
}
