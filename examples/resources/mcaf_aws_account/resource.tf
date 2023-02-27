provider "mcaf" {
  aws {
    region = "eu-west-1"
  }
}

# resource "random_string" "example_nested_ou" {
#   length  = 5
#   special = false
# }

resource "mcaf_aws_account" "example_nested_ou" {
  name                = "example-nested-ou-test-closing"
  email               = "example-nested-ou-test@example.com"
  organizational_unit_path = "Custom/Test1"
  close_on_deletion = true
  
  sso {
    firstname = "Example"
    lastname  = "Admin"
    email     = "example-admin@example.com"
  }
}
