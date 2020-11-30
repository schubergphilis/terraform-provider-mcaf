package mcaf

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceO365Alias() *schema.Resource {
	return &schema.Resource{
		Create: checkProvider("o365", resourceO365AliasCreate),
		Read:   checkProvider("o365", resourceO365AliasRead),
		Delete: checkProvider("o365", resourceO365AliasDelete),
		Importer: &schema.ResourceImporter{
			State: resourceO365AliasImporter,
		},

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("O365_GROUP_ID", nil),
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (warns []string, errs []error) {
					if _, err := uuid.Parse(v.(string)); err != nil {
						errs = append(errs, fmt.Errorf("%q is not a valid UUID", k))
					}
					return warns, errs
				},
			},
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceO365AliasResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceO365AliasStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceO365AliasCreate(d *schema.ResourceData, meta interface{}) error {
	o365 := meta.(*Client).O365

	// Get the alias and group ID.
	alias := d.Get("alias").(string)
	groupID := d.Get("group_id").(string)

	log.Printf("[DEBUG] Add alias %s to group %s", alias, groupID)
	_, err := o365.Create(groupID, alias)
	if err != nil {
		return fmt.Errorf("Error creating alias %s: %v", alias, err)
	}

	// Set the ID.
	d.SetId(alias)

	return nil
}

func resourceO365AliasRead(d *schema.ResourceData, meta interface{}) error {
	o365 := meta.(*Client).O365

	// Get the alias and group ID.
	alias := d.Id()
	groupID := d.Get("group_id").(string)

	log.Printf("[DEBUG] Read existing aliases of group %s", groupID)
	resp, err := o365.Read(groupID)
	if err != nil {
		return fmt.Errorf("Error reading aliases: %v", err)
	}

	// Check and return if the alias exists.
	for _, nickname := range resp.Group.Aliases {
		if alias == nickname {
			d.SetId(alias)
			d.Set("alias", alias)
			return nil
		}
	}

	// Unset the ID as the alias no longer exists.
	d.SetId("")

	return nil
}

func resourceO365AliasDelete(d *schema.ResourceData, meta interface{}) error {
	o365 := meta.(*Client).O365

	// Get the alias and group ID.
	alias := d.Id()
	groupID := d.Get("group_id").(string)

	log.Printf("[DEBUG] Delete alias %s from group %s", alias, groupID)
	_, err := o365.Delete(groupID, alias)
	if err != nil {
		return fmt.Errorf("Error deleting alias %s: %v", alias, err)
	}

	return nil
}

func resourceO365AliasImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf(
			"invalid O365 alias import format: %s (expected <GROUP ID>:<ALIAS>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("group_id", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}
