package mcaf

import (
	"log"

	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	homedir "github.com/mitchellh/go-homedir"
)

// Client represents a general purpose MCAF client.
type Client struct {
	AWSClient *AWSClient
}

type AWSClient struct {
	accountID string
	cbconn    *codebuild.CodeBuild
	scconn    *servicecatalog.ServiceCatalog
}

// awsClient configures and returns a fully initialized AWSClient.
func awsClient(aws map[string]interface{}) (*AWSClient, error) {
	log.Println("[INFO] Building AWS auth structure")
	config := &awsbase.Config{
		AccessKey:               aws["access_key"].(string),
		SecretKey:               aws["secret_key"].(string),
		Profile:                 aws["profile"].(string),
		Region:                  aws["region"].(string),
		MaxRetries:              aws["max_retries"].(int),
		SkipCredsValidation:     aws["skip_credentials_validation"].(bool),
		SkipMetadataApiCheck:    aws["skip_metadata_api_check"].(bool),
		SkipRequestingAccountId: aws["skip_requesting_account_id"].(bool),
		Token:                   aws["token"].(string),
		DebugLogging:            logging.IsDebugOrHigher(),
		UserAgentProducts: []*awsbase.UserAgentProduct{
			{Name: "MCAF", Version: "1.0"},
			{Name: "Terraform", Version: terraform.NewState().TFVersion},
		},
	}

	// Set CredsFilename, expanding home directory
	credsPath, err := homedir.Expand(aws["shared_credentials_file"].(string))
	if err != nil {
		return nil, err
	}
	config.CredsFilename = credsPath

	sess, accountID, _, err := awsbase.GetSessionWithAccountIDAndPartition(config)
	if err != nil {
		return nil, err
	}

	if accountID == "" {
		log.Printf("[WARN] AWS account ID not found for provider")
	}

	client := &AWSClient{
		accountID: accountID,
		cbconn:    codebuild.New(sess.Copy()),
		scconn:    servicecatalog.New(sess.Copy()),
	}

	return client, nil
}
