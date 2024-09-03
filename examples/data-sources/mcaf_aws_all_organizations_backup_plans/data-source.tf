provider "mcaf" {
  aws {}
}

data "mcaf_aws_all_organization_backup_configuration" "example" {}

output "mcaf_aws_all_organization_backup_configuration" {
  value = data.mcaf_aws_all_organization_backup_configuration.example
}
