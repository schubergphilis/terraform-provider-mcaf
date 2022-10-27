package mcaf

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type OrganizationalUnit struct {
	OrganizationalUnit *organizations.OrganizationalUnit

	Path *string
}

func dataSourceAwsAllOrganizationalUnits() *schema.Resource {
	return &schema.Resource{
		Read: checkProvider("aws", dataSourceAwsAllOrganizationalUnitsRead),

		Schema: map[string]*schema.Schema{
			"organizational_units": {
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
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsAllOrganizationalUnitsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*Client).AWSClient.orgsconn

	roots, err := listRoots(conn)
	if err != nil {
		return err
	}
	root_id := aws.StringValue(roots[0].Id)

	var ous []*OrganizationalUnit
	ous, err = listOrganizationalUnitsForParentPagesRecursive(conn, "Root", root_id, ous)
	if err != nil {
		return err
	}

	d.SetId(root_id)

	if err := d.Set("organizational_units", flattenOrganizationsOrganizationalUnits(ous)); err != nil {
		return fmt.Errorf("Error setting organizational_units: %s", err)
	}

	return nil
}

func flattenOrganizationsOrganizationalUnits(ous []*OrganizationalUnit) []map[string]interface{} {
	if len(ous) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, ou := range ous {
		result = append(result, map[string]interface{}{
			"arn":  aws.StringValue(ou.OrganizationalUnit.Arn),
			"id":   aws.StringValue(ou.OrganizationalUnit.Id),
			"name": aws.StringValue(ou.OrganizationalUnit.Name),
			"path": aws.StringValue(ou.Path),
		})
	}
	return result
}

func listOrganizationalUnitsForParentPagesRecursive(conn *organizations.Organizations, parentPath, parentId string, ous []*OrganizationalUnit) ([]*OrganizationalUnit, error) {
	// Control Tower supports a maximum of 5 levels of nested OUs.
	parentPathSplit := strings.Split(parentPath, "/")
	if len(parentPathSplit) == 5 {
		log.Printf("[INFO] Maximum number of levels of nested OUs reached. Skipping %s (%s)", parentPath, parentId)
		return ous, nil
	}

	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parentId),
	}

	log.Printf("[DEBUG] Listing OUs under parent: %s (%s)", parentPath, parentId)
	err := conn.ListOrganizationalUnitsForParentPages(input, func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
		for _, ou := range page.OrganizationalUnits {
			ouPath := fmt.Sprintf("%s/%s", parentPath, aws.StringValue(ou.Name))
			ous = append(ous, &OrganizationalUnit{
				OrganizationalUnit: ou,
				Path:               aws.String(ouPath),
			})

			var err error
			ous, err = listOrganizationalUnitsForParentPagesRecursive(conn, ouPath, aws.StringValue(ou.Id), ous)
			if err != nil {
				log.Printf("[ERROR] Error listing OUs for %s (%s): %s", ouPath, aws.StringValue(ou.Id), err)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("error listing Organization Units for parent (%s): %s", parentId, err)
	}

	return ous, nil
}

func listRoots(conn *organizations.Organizations) ([]*organizations.Root, error) {
	var roots []*organizations.Root
	err := conn.ListRootsPages(&organizations.ListRootsInput{}, func(page *organizations.ListRootsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		roots = append(roots, page.Roots...)

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("error listing Organization Roots: %s", err)
	}

	return roots, nil
}
