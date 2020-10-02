package mcaf

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	O365_ALIAS    = os.Getenv("O365_ALIAS")
	O365_GROUP_ID = os.Getenv("O365_GROUP_ID")
)

func TestAccMcafO365Alias_basic(t *testing.T) {

	var group o365Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMcafO365AliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMcafO365Alias_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMcafO365AliasExists(
						"mcaf_o365_alias.foo", &group),
					testAccMcafO365AliasAttributes(&group),
					resource.TestCheckResourceAttr(
						"mcaf_o365_alias.foo", "alias", O365_ALIAS),
					resource.TestCheckResourceAttr(
						"mcaf_o365_alias.foo", "group_id", O365_GROUP_ID),
					resource.TestCheckResourceAttr(
						"mcaf_o365_alias.foo", "id", O365_ALIAS),
				),
			},

			{
				ResourceName:      "mcaf_o365_alias.foo",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s:%s", O365_GROUP_ID, O365_ALIAS),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMcafO365AliasExists(n string, group *o365Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No alias ID is set")
		}

		client := testAccProvider.Meta().(*Client).O365
		g, err := client.Read(rs.Primary.Attributes["group_id"])
		if err != nil {
			return err
		}

		if !contains(g.Group.Aliases, O365_ALIAS) {
			return fmt.Errorf("Alias not found")
		}

		*group = g.Group

		return nil
	}
}

func testAccMcafO365AliasAttributes(group *o365Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if group.Id != O365_GROUP_ID {
			return fmt.Errorf("Bad group ID: %s", group.Id)
		}

		return nil
	}
}

func testAccMcafO365AliasDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).O365

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mcaf_o365_alias" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No group ID is set")
		}

		_, err := client.Read(rs.Primary.Attributes["group_id"])
		if err == nil {
			return fmt.Errorf("Alias %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

var testAccMcafO365Alias_basic = fmt.Sprintf(`
provider "mcaf" {
  o365 {}
}

resource "mcaf_o365_alias" "foo" {
  alias = "%s"
}`,
	O365_ALIAS,
)

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
