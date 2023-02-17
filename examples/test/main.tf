
resource "mcaf_aws_account" "example" {
  name                     = "testworkload-sbx"
  email                    = "int+testworkload-sbx@aws.devolksbank.nl"
  organizational_unit_path = "Workload/Sandbox"
  sso {
    firstname = "Control Tower"
    lastname  = "Admin"
    email     = "aws@devolksbank.nl"
  }
  lifecycle {
    prevent_destroy = true
  }
}