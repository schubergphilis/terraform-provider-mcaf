provider "mcaf" {
  aws {}
}

data "mcaf_aws_all_organizational_units" "example" {}

output "mcaf_aws_all_organizational_units" {
  value = data.mcaf_aws_all_organizational_units.example.organizational_units
}
