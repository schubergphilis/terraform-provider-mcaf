package mcaf

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type OrganizationBackupConfiguration struct {
	Backup *backup.Backup
}

type OrganizationBackupPlan struct {
	plan       *backup.PlansListMember
	selections []*OrganizationBackupSelection
}

type OrganizationBackupSelection struct {
	selection *backup.SelectionsListMember
	tags      map[string]string
}

func dataSourceAwsAllOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: checkProvider("aws", dataSourceAwsAllOrganizationConfigurationRead),

		Schema: map[string]*schema.Schema{
			"organization_backup_plans": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_selections": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"tags": {
										Type:     schema.TypeMap,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsAllOrganizationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*Client).AWSClient.bconn

	backupPlans, err := listBackupPlansFromOrgPolicies(conn)
	if err != nil {
		return err
	}

	d.SetId(meta.(*Client).AWSClient.accountID + "/org-backup-plans")

	if err := d.Set("organization_backup_plans", flattenOrganizationsBackupPlans(backupPlans)); err != nil {
		return fmt.Errorf("Error setting organization_backup_plans: %s", err)
	}

	return nil
}

func flattenOrganizationsBackupPlans(backupPlans []*OrganizationBackupPlan) []map[string]interface{} {
	if len(backupPlans) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, backupPlan := range backupPlans {
		result = append(result, map[string]interface{}{
			"arn":               aws.StringValue(backupPlan.plan.BackupPlanArn),
			"id":                aws.StringValue(backupPlan.plan.BackupPlanId),
			"name":              aws.StringValue(backupPlan.plan.BackupPlanName),
			"backup_selections": flattenOrganizationsBackupSelections(backupPlan.selections),
		})
	}
	return result
}

func flattenOrganizationsBackupSelections(backupSelections []*OrganizationBackupSelection) []map[string]interface{} {
	if len(backupSelections) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, backupSelection := range backupSelections {
		result = append(result, map[string]interface{}{
			"id":   aws.StringValue(backupSelection.selection.SelectionId),
			"tags": aws.StringMap(backupSelection.tags),
		})
	}
	return result
}

func listBackupPlansFromOrgPolicies(conn *backup.Backup) ([]*OrganizationBackupPlan, error) {

	var backupPlans []*OrganizationBackupPlan
	err := conn.ListBackupPlansPages(&backup.ListBackupPlansInput{}, func(page *backup.ListBackupPlansOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, s := range page.BackupPlansList {
			if strings.HasPrefix(*s.BackupPlanId, "orgs/") {
				backupSelections, err := listBackupSelectionsWithBackupPlanId(conn, s.BackupPlanId)
				if err != nil {
					backupSelections = make([]*OrganizationBackupSelection, 0)
				}

				backupPlans = append(backupPlans, &OrganizationBackupPlan{
					plan:       s,
					selections: backupSelections,
				})
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("error listing backup plans: %s", err)
	}

	return backupPlans, nil
}

func listBackupSelectionsWithBackupPlanId(conn *backup.Backup, backupPlanId *string) ([]*OrganizationBackupSelection, error) {

	var backupSelections []*OrganizationBackupSelection
	params := &backup.ListBackupSelectionsInput{
		BackupPlanId: backupPlanId,
	}
	err := conn.ListBackupSelectionsPages(params, func(page *backup.ListBackupSelectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, s := range page.BackupSelectionsList {

			tags, err := listBackupSelectionsTagsWithBackupSelection(conn, s)
			if err != nil {
				tags = nil
			}
			backupSelections = append(backupSelections, &OrganizationBackupSelection{
				selection: s,
				tags:      tags,
			})
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("error listing backup selections: %s", err)
	}

	return backupSelections, nil
}

func listBackupSelectionsTagsWithBackupSelection(conn *backup.Backup, backupSelectionListMember *backup.SelectionsListMember) (map[string]string, error) {

	tags := make(map[string]string)
	params := &backup.GetBackupSelectionInput{
		BackupPlanId: backupSelectionListMember.BackupPlanId,
		SelectionId:  backupSelectionListMember.SelectionId,
	}
	backupSelection, err := conn.GetBackupSelection(params)

	if err != nil {
		return nil, fmt.Errorf("error listing backup selections: %s", err)
	}

	for _, s := range backupSelection.BackupSelection.ListOfTags {
		tags[*s.ConditionKey] = *s.ConditionValue
	}

	return tags, nil
}
