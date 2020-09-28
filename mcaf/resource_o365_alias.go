package mcaf

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceO365Alias() *schema.Resource {
	return &schema.Resource{
		Create: checkProvider("o365", resourceO365AliasCreate),
		Read:   checkProvider("o365", resourceO365AliasRead),
		Delete: checkProvider("o365", resourceO365AliasDelete),

		Schema: map[string]*schema.Schema{
			"alias": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"group_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("O365_GROUP_ID", nil),
				ForceNew:    true,
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
	_, err := o365.Create(groupID, []string{alias})
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

	// Check and return if the alias still exists.
	for _, nickname := range resp.Aliases {
		if alias == nickname {
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
	_, err := o365.Delete(groupID, []string{alias})
	if err != nil {
		return fmt.Errorf("Error deleting alias %s: %v", alias, err)
	}

	return nil
}
