package mcaf

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsAllOrganizationalUnits_basic(t *testing.T) {
	dataSourceName := "data.mcaf_aws_all_organizational_units.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccAwsPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAllOrganizationalUnitsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "organizational_units.#", "6"),
				),
			},
		},
	})
}

const testAccDataSourceAwsAllOrganizationalUnitsConfig = `
provider "mcaf" {
  aws {}
}

data "mcaf_aws_all_organizational_units" "test" {}
`
