package mcaf

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAWSAccount() *schema.Resource {
	return &schema.Resource{
		Create: checkProvider("aws", resourceAWSAccountCreate),
		Read:   checkProvider("aws", resourceAWSAccountRead),
		Update: checkProvider("aws", resourceAWSAccountUpdate),
		Delete: checkProvider("aws", resourceAWSAccountDelete),

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"close_on_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organizational_unit": {
				Type:          schema.TypeString,
				Optional:      true,
				Deprecated:    "This field is deprecated and will be removed in a future version. Please use organizational_unit_path instead.",
				ConflictsWith: []string{"organizational_unit_path"},
			},
			"organizational_unit_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"organizational_unit"},
			},
			"provisioned_product_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"sso": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"firstname": {
							Type:     schema.TypeString,
							Required: true,
						},

						"lastname": {
							Type:     schema.TypeString,
							Required: true,
						},

						"email": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

var accountMutex sync.Mutex

func resourceAWSAccountCreate(d *schema.ResourceData, meta interface{}) error {
	orgsconn := meta.(*Client).AWSClient.orgsconn
	scconn := meta.(*Client).AWSClient.scconn

	log.Printf("[DEBUG] Search the Account Factory product")
	products, err := scconn.SearchProducts(&servicecatalog.SearchProductsInput{
		Filters: map[string][]*string{"FullTextSearch": {aws.String("AWS Control Tower Account Factory")}},
	})
	if err != nil {
		return fmt.Errorf("Error searching service catalog: %v", err)
	}
	if len(products.ProductViewSummaries) != 1 {
		return fmt.Errorf("No Control Tower Account Factory found in your account. Please check your User and/or Role permissions and your Region settings.")
	}

	log.Printf("[DEBUG List all product artifacts to find the active artifact")
	artifacts, err := scconn.ListProvisioningArtifacts(&servicecatalog.ListProvisioningArtifactsInput{
		ProductId: products.ProductViewSummaries[0].ProductId,
	})
	if err != nil {
		return fmt.Errorf("Error listing provisioning artifacts: %v", err)
	}

	// Try to find the active (which should be the latest) artifact.
	artifactID := ""
	for _, artifact := range artifacts.ProvisioningArtifactDetails {
		if *artifact.Active {
			artifactID = *artifact.Id
			break
		}
	}
	if artifactID == "" {
		return fmt.Errorf("Could not find the provisioning artifact ID")
	}

	// Get organisation Root OU name and ID
	roots, err := listRoots(orgsconn)
	if err != nil {
		return err
	}

	// Support both organizational_unit and organizational_unit_path until deprecated organizational_unit field is removed
	ou, ouOk := d.GetOk("organizational_unit")
	ouPath, ouPathOk := d.GetOk("organizational_unit_path")
	if !ouOk && !ouPathOk {
		return errors.New("one of organizational_unit or organizational_unit_path must be configured")
	}
	if ouOk {
		ouPath = ou
	}

	// Get child OU name and ID from provided path
	managedOu, err := returnChildOu(orgsconn, ouPath.(string), aws.StringValue(roots[0].Id), aws.StringValue(roots[0].Name))
	if err != nil {
		return err
	}

	// Get the name, ou and SSO details from the config.
	name := d.Get("name").(string)
	ppn := d.Get("provisioned_product_name").(string)
	sso := d.Get("sso").([]interface{})[0].(map[string]interface{})

	// If no provisioned product name was configured, use the name.
	if ppn == "" {
		ppn = name
	}

	// Create a new parameters struct.
	params := &servicecatalog.ProvisionProductInput{
		ProductId:              products.ProductViewSummaries[0].ProductId,
		ProvisionedProductName: aws.String(ppn),
		ProvisioningArtifactId: aws.String(artifactID),
		ProvisioningParameters: []*servicecatalog.ProvisioningParameter{
			{
				Key:   aws.String("AccountName"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("AccountEmail"),
				Value: aws.String(d.Get("email").(string)),
			},
			{
				Key:   aws.String("SSOUserFirstName"),
				Value: aws.String(sso["firstname"].(string)),
			},
			{
				Key:   aws.String("SSOUserLastName"),
				Value: aws.String(sso["lastname"].(string)),
			},
			{
				Key:   aws.String("SSOUserEmail"),
				Value: aws.String(sso["email"].(string)),
			},
			{
				Key:   aws.String("ManagedOrganizationalUnit"),
				Value: aws.String(fmt.Sprintf("%s (%s)", aws.StringValue(managedOu.Name), aws.StringValue(managedOu.Id))),
			},
		},
	}

	log.Printf("[DEBUG] Provision product parameters: %+v\n", params)

	accountMutex.Lock()
	defer accountMutex.Unlock()

	log.Printf("[DEBUG] Provision account %s in organizational unit %s (%s)", name, aws.StringValue(managedOu.Name), aws.StringValue(managedOu.Id))
	account, err := scconn.ProvisionProduct(params)
	if err != nil {
		return fmt.Errorf("Error provisioning account %s: %v", name, err)
	}

	// Set the ID so we can cleanup the provisioned account in case of a failure.
	d.SetId(*account.RecordDetail.ProvisionedProductId)

	// Wait for the provisioning to finish.
	err = waitForProvisioning(name, account.RecordDetail.RecordId, meta)
	if err != nil {
		return err
	}

	return resourceAWSAccountRead(d, meta)
}

func resourceAWSAccountRead(d *schema.ResourceData, meta interface{}) error {
	orgsconn := meta.(*Client).AWSClient.orgsconn
	scconn := meta.(*Client).AWSClient.scconn

	// Get the name from the config.
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Read configuration of provisioned account %s: %s", name, d.Id())
	account, err := scconn.DescribeProvisionedProduct(&servicecatalog.DescribeProvisionedProductInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error reading configuration of provisioned account %s: %v", name, err)
	}

	record := &servicecatalog.DescribeRecordInput{
		Id: account.ProvisionedProductDetail.LastRecordId,
	}

	status, err := scconn.DescribeRecord(record)
	if err != nil {
		return fmt.Errorf("Error reading configuration of provisioned account %s: %v", name, err)
	}

	// Update the config.
	d.Set("provisioned_product_name", *account.ProvisionedProductDetail.Name)
	var accountID string
	for _, output := range status.RecordOutputs {
		switch *output.OutputKey {
		case "AccountName":
			d.Set("name", *output.OutputValue)
		case "AccountEmail":
			d.Set("email", *output.OutputValue)
		case "AccountId":
			accountID = *aws.String(*output.OutputValue)
			d.Set("account_id", *output.OutputValue)
		}
	}

	accountInfo, err := getAccountInfo(orgsconn, accountID)
	if err != nil {
		return fmt.Errorf("error reading AWS Organizations Account (%s): %w", accountID, err)
	}

	d.Set("organization_account_status", accountInfo.Status)
	return nil
}

func resourceAWSAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	orgsconn := meta.(*Client).AWSClient.orgsconn
	scconn := meta.(*Client).AWSClient.scconn

	// Get organisation Root OU name and ID
	roots, err := listRoots(orgsconn)
	if err != nil {
		return err
	}

	// Support both organizational_unit and organizational_unit_path until deprecated organizational_unit field is removed
	ou, ouOk := d.GetOk("organizational_unit")
	ouPath, ouPathOk := d.GetOk("organizational_unit_path")
	if !ouOk && !ouPathOk {
		return errors.New("one of organizational_unit or organizational_unit_path must be configured")
	}
	if ouOk {
		ouPath = ou
	}

	// Get child OU name and ID from provided path
	managedOu, err := returnChildOu(orgsconn, ouPath.(string), aws.StringValue(roots[0].Id), aws.StringValue(roots[0].Name))
	if err != nil {
		return err
	}

	// Get the name, ou and SSO details from the config.
	name := d.Get("name").(string)
	sso := d.Get("sso").([]interface{})[0].(map[string]interface{})

	// Create a new parameters struct.
	params := &servicecatalog.UpdateProvisionedProductInput{
		ProvisionedProductId: aws.String(d.Id()),
		ProvisioningParameters: []*servicecatalog.UpdateProvisioningParameter{
			{
				Key:   aws.String("AccountName"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("AccountEmail"),
				Value: aws.String(d.Get("email").(string)),
			},
			{
				Key:   aws.String("SSOUserFirstName"),
				Value: aws.String(sso["firstname"].(string)),
			},
			{
				Key:   aws.String("SSOUserLastName"),
				Value: aws.String(sso["lastname"].(string)),
			},
			{
				Key:   aws.String("SSOUserEmail"),
				Value: aws.String(sso["email"].(string)),
			},
			{
				Key:   aws.String("ManagedOrganizationalUnit"),
				Value: aws.String(fmt.Sprintf("%s (%s)", aws.StringValue(managedOu.Name), aws.StringValue(managedOu.Id))),
			},
		},
	}

	accountMutex.Lock()
	defer accountMutex.Unlock()

	log.Printf("[DEBUG] Update provisioned account %s: %s", name, d.Id())
	account, err := scconn.UpdateProvisionedProduct(params)
	if err != nil {
		return fmt.Errorf("Error updating provisioned account %s: %v", name, err)
	}

	// Wait for the provisioning to finish.
	err = waitForProvisioning(name, account.RecordDetail.RecordId, meta)
	if err != nil {
		return err
	}

	return resourceAWSAccountRead(d, meta)
}

func resourceAWSAccountDelete(d *schema.ResourceData, meta interface{}) error {
	scconn := meta.(*Client).AWSClient.scconn
	orgconn := meta.(*Client).AWSClient.orgsconn

	// Get the name from the config.
	name := d.Get("name").(string)
	close := d.Get("close_on_deletion").(bool)

	accountMutex.Lock()
	defer accountMutex.Unlock()

	log.Printf("[DEBUG] Deleting provisioned account %s: %s", name, d.Id())
	account, err := scconn.TerminateProvisionedProduct(&servicecatalog.TerminateProvisionedProductInput{
		ProvisionedProductId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting provisioned account %s: %v", name, err)
	}

	// Wait for the provisioning to finish.
	err = waitForProvisioning(name, account.RecordDetail.RecordId, meta)
	if err != nil {
		return fmt.Errorf("Error waiting for deletion provisioned account %s: %v", name, err)
	}

	if close {
		accountID := getAccountID(scconn, account.RecordDetail.RecordId)
		log.Printf("[DEBUG] Closing AWS Organizations Account: %s", accountID)
		_, err = orgconn.CloseAccount(&organizations.CloseAccountInput{
			AccountId: aws.String(accountID),
		})
		if err != nil {
			return fmt.Errorf("Error deleting AWS Organizations Account (%s): %w", accountID, err)
		}
		if _, err := waitForDeleting(orgconn, accountID); err != nil {
			return fmt.Errorf("Error waiting for AWS Organizations Account (%s) delete: %w", accountID, err)
		}
	}

	return nil
}

// returnChildOu returns the ID of the child OU with the given path.
func returnChildOu(conn *organizations.Organizations, path, ouID, ouName string) (*organizations.OrganizationalUnit, error) {
	ou := &organizations.OrganizationalUnit{}

	for _, v := range strings.Split(path, "/") {
		if strings.EqualFold(v, "Root") {
			continue
		}

		input := &organizations.ListOrganizationalUnitsForParentInput{
			ParentId: aws.String(ouID),
		}

		var childOuID, childOuName string
		log.Printf("[DEBUG] Listing OUs under parent: %s (%s)", ouName, ouID)
		err := conn.ListOrganizationalUnitsForParentPages(input, func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
			for _, childOu := range page.OrganizationalUnits {
				if *childOu.Name == v {
					childOuID = *childOu.Id
					childOuName = *childOu.Name
					ou = childOu
					return false
				}
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("error listing organizational units for parent %s (%s): %v", ouName, ouID, err)
		}

		if childOuID == "" {
			return nil, fmt.Errorf("organizational unit %s not found in parent %s (%s)", ou, ouName, ouID)
		}

		ouID = childOuID
		ouName = childOuName
	}

	return ou, nil
}

// waitForProvisioning waits until the provisioning finished.
func waitForProvisioning(name string, recordID *string, meta interface{}) error {
	scconn := meta.(*Client).AWSClient.scconn

	record := &servicecatalog.DescribeRecordInput{
		Id: recordID,
	}

	for {
		// Get the provisioning status.
		status, err := scconn.DescribeRecord(record)
		if err != nil {
			return fmt.Errorf("Error reading provisioning status of account %s: %v", name, err)
		}

		// If the provisioning succeeded we are done.
		if *status.RecordDetail.Status == servicecatalog.RecordStatusSucceeded {
			break
		}

		// If the provisioning failed we try to cleanup the tainted account.
		if *status.RecordDetail.Status == servicecatalog.RecordStatusFailed {
			return fmt.Errorf("Provisioning account %s failed: %s", name, *status.RecordDetail.RecordErrors[0].Description)
		}

		// Wait 5 seconds before checking the status again.
		time.Sleep(5 * time.Second)
	}

	return nil
}

// waitForDeleting waits until account be suspended
func waitForDeleting(orgconn *organizations.Organizations, id string) (*organizations.Account, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{organizations.AccountStatusPendingClosure},
		Target:       []string{organizations.AccountStatusSuspended},
		Refresh:      getAccountStateStatus(orgconn, id),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*organizations.Account); ok {
		return output, err
	}

	return nil, err
}

// getAccountID gets the account id based on rercord from the termination of the provisioned product
func getAccountID(svc *servicecatalog.ServiceCatalog, recordID *string) string {
	var accountId string

	record := &servicecatalog.DescribeRecordInput{
		Id: recordID,
	}

	describeRecord, err := svc.DescribeRecord(record)
	if err != nil {
		fmt.Print(err)
	}
	for _, output := range describeRecord.RecordOutputs {
		if *output.OutputKey == "AccountId" {
			accountId = *output.OutputValue
			break
		}
	}

	return accountId
}

// getAccountStateStatus gets account ID and returns StateRefreshFunc used to watch account resource state status
func getAccountStateStatus(orgconn *organizations.Organizations, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := getAccountInfo(orgconn, id)
		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// getAccountInfo gets account ID and returns account info
func getAccountInfo(conn *organizations.Organizations, id string) (*organizations.Account, error) {
	input := &organizations.DescribeAccountInput{
		AccountId: aws.String(id),
	}

	output, err := conn.DescribeAccount(input)
	if err != nil {
		return nil, err
	}

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAccountNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	return output.Account, nil
}
