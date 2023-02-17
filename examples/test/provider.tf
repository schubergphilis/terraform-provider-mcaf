terraform {
  required_providers {
    mcaf = {
      source = "terraform-provider-mcaf/local/mcaf"
    }
  }
}

provider "mcaf" {
  aws {

  }
}