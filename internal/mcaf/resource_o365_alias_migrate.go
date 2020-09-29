package mcaf

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceO365AliasResourceV0() *schema.Resource {
	return &schema.Resource{
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
			},
		},
	}
}

func resourceO365AliasStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {

	log.Println("[INFO] Found O365 group alias state v0; migrating to v1")

	group_id := rawState["group_id"].(string)
	uuid := strings.Split(group_id, "_")[1]
	rawState["group_id"] = uuid

	return rawState, nil
}
