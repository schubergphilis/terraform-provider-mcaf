package mcaf

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// New returns a schema.Provider.
func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"aws": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     awsProviderSchema(),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"mcaf_aws_all_organizational_units":              dataSourceAwsAllOrganizationalUnits(),
			"mcaf_aws_all_organization_backup_configuration": dataSourceAwsAllOrganizationConfiguration(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"mcaf_aws_account":           resourceAWSAccount(),
			"mcaf_aws_codebuild_trigger": resourceAWSCodeBuildTrigger(),
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
