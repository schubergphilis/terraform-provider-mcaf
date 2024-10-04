---
layout: "mcaf"
page_title: "Provider: MCAF"
sidebar_current: "docs-mcaf-index"
description: |-
  The MCAF provider is used to provide additional functionality not available in official or community providers.
---

# MCAF Provider

The MCAF provider is used to provide additional functionality not available in official or community providers.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "mcaf" {
  aws {}
}
```

## Argument Reference

Refer to the <a href="https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication">AWS
Provider authentication docs</a> for how to configure the `aws` object. We
recommend exporting the following variables:

* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `AWS_DEFAULT_REGION`
