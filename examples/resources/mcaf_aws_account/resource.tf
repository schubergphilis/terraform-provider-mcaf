provider "mcaf" {
  aws {}
}

resource "random_string" "example_top_ou" {
  length  = 5
  special = false
}

resource "mcaf_aws_account" "example_top_ou" {
  name                = "example-top-ou-${random_string.example_top_ou.result}"
  email               = "example-top-ou-${random_string.example_top_ou.result}@example.com"
  organizational_unit = "Custom"

  sso {
    firstname = "Example"
    lastname  = "Admin"
    email     = "example-admin@example.com"
  }
}

resource "random_string" "example_nested_ou" {
  length  = 5
  special = false
}

resource "mcaf_aws_account" "example_nested_ou" {
  name                = "example-nested-ou-${random_string.example_nested_ou.result}"
  email               = "example-nested-ou-${random_string.example_nested_ou.result}@example.com"
  organizational_unit = "Custom/Test1"

  sso {
    firstname = "Example"
    lastname  = "Admin"
    email     = "example-admin@example.com"
  }
}