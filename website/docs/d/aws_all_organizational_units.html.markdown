---
layout: "mcaf"
page_title: "MCAF: mcaf_aws_all_organizational_units"
sidebar_current: "docs-datasource-mcaf-aws-all-organizational-units"
description: |-
  Recursively get all organizational units under the Root organizational unit.
---

# Data Source: mcaf_aws_all_organizational_units

Recursively get all organizational units under the Root organizational unit.

## Example Usage

```hcl
data "mcaf_aws_all_organizational_units" "example" {}
```

## Attributes Reference

* `organizational_units` - List of child organizational units, which have the following attributes:
  * `arn` - ARN of the organizational unit
  * `name` - Name of the organizational unit
  * `id` - ID of the organizational unit
  * `path` - Full path of the organizational unit, e.g. `Root/Core`
* `id` - ID of the "Root" organizational unit
