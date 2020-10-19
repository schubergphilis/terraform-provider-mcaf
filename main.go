package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/schubergphilis/terraform-provider-mcaf/internal/mcaf"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mcaf.New,
	})
}
