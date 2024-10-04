---
layout: "mcaf"
page_title: "MCAF: mcaf_aws_account"
sidebar_current: "docs-mcaf-resource-aws-account"
description: |-
  Creates an AWS account using Control Tower's Account Factory.
---

# mcaf_aws_account

Creates an AWS account using Control Tower's Account Factory.

## Example Usage

```hcl
resource "mcaf_aws_account" "example" {
  name                     = "foo"
  email                    = "foo@example"
  organizational_unit_path = "My-OU"

  sso {
    firstname = "Control Tower"
    lastname  = "Admin"
    email     = "control-tower@example.com"
  }
}
```

~> **NOTE:** Deleting a `mcaf_aws_account` resource does not close the account. Instead, the provisioned product is deleted resulting in account being moved to the Root OU and un-enrolled from Control Tower. Closing the account as part of the deletion will be handled in a future version.

It is also possible to create an AWS account in a nested organizational unit by specifying it's path:

```hcl
resource "mcaf_aws_account" "example" {
  name                     = "foo"
  email                    = "foo@example"
  organizational_unit_path = "My-Team/My-Project/My-Env"

  sso {
    firstname = "Control Tower"
    lastname  = "Admin"
    email     = "control-tower@example.com"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the account.

- `email` - (Required) The email address of the account.

- `organizational_unit` - (Optional) The Organizational Unit to place the account in. **Deprecated** This argument has been replaced by `organizational_unit_path` and will be removed in a future version.

- `organizational_unit_path` - (Optional) The Organizational Unit path to place the account in.

- `provisioned_product_path_id` - (Optional) The launch path ID of the `AWS Control Tower Account Factory` product. If not provided, the provider will attempt to look it up itself.

- `provisioned_product_name` - (Optional) A custom name for the provisioned product.

The `sso` object supports the following:

- `firstname` - (Required) The first name of the Control Tower SSO account.

- `lastname` - (Required) The lastname of the Control Tower SSO account.

- `email` - (Required) The email address of the Control Tower SSO account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `account_id` - The ID of the AWS account.
