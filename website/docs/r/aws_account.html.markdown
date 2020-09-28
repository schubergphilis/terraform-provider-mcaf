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
  name                = "foo"
  email               = "foo@example"
  organizational_unit = "My-OU"

  sso {
    firstname = "Control Tower"
    lastname  = "Admin"
    email     = "control-tower@example.com"
  }

  lifecycle {
    prevent_destroy = true
  }
}
```

~> **NOTE:** You cannot delete an AWS account so the lifecycle block required. Sometime in the future the delete action may 

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the account.

* `email` - (Required) The email address of the account.

* `organizational_unit` - (Required) The Organizational Unit to place the account in.

* `provisioned_product_name` - (Optional) A custom name for the provisioned product.

The `sso` object supports the following:

* `firstname` - (Required) The first name of the Control Tower SSO account.

* `lastname` - (Required) The lastname of the Control Tower SSO account.

* `email` - (Required) The email address of the Control Tower SSO account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `account_id` - The ID of the AWS account.
