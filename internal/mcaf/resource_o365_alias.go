package mcaf

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceO365Alias() *schema.Resource {
	return &schema.Resource{
		Create: checkProvider("o365", resourceO365AliasCreate),
		Read:   checkProvider("o365", resourceO365AliasRead),
		Delete: checkProvider("o365", resourceO365AliasDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
					if !isValidUUID(v.(string)) {
						errs = append(errs, fmt.Errorf("%q is not a valid UUID", k))
					}
					return
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

	// Sleep to compensate for delay in Exchange Online API to show new aliases.
	time.Sleep(10 * time.Second)

	return nil
}

func resourceO365AliasRead(d *schema.ResourceData, meta interface{}) error {
	o365 := meta.(*Client).O365

	// Get the alias and group ID.
	alias := d.Id()
	groupID := d.Get("group_id").(string)

	// Incase we're importing a resource...
	if strings.Contains(d.Id(), ":") {
		var err error
		groupID, alias, err = resourceO365AliasParseId(d.Id())
		if err != nil {
			return err
		}
		d.SetId(alias)
		d.Set("group_id", groupID)
	}

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

func resourceO365AliasParseId(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected groupid:alias", id)
	}

	return parts[0], parts[1], nil
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
