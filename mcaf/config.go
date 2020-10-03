package mcaf

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	homedir "github.com/mitchellh/go-homedir"
)

// Client represents a general purpose MCAF client.
type Client struct {
	AWSClient *AWSClient
	O365      *O365Client
}

type AWSClient struct {
	accountID string
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
			{Name: "APN", Version: "1.0"},
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
		scconn:    servicecatalog.New(sess.Copy()),
	}

	return client, nil
}

// o365Client configures and returns a fully initialized O365Client.
func o365Client(o365 map[string]interface{}) (*O365Client, error) {
	u, err := url.Parse(o365["endpoint"].(string))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse endpoint: %v", err)
	}

	return &O365Client{
		aclGUID:    o365["acl_guid"].(string),
		client:     &http.Client{Timeout: 10 * time.Minute},
		endpoint:   u,
		secretCode: o365["secret_code"].(string),
	}, nil
}
