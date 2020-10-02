---
layout: "mcaf"
page_title: "MCAF: mcaf_o365_alias"
sidebar_current: "docs-mcaf-resource-o365-alias"
description: |-
  Adds a proxy address to an existing O365 group using the ExoAPI.
---

# mcaf_o365_alias

Adds a ProxyAddress to an existing O365 group using the ExoAPI.

## Example Usage

```hcl
resource "mcaf_o365_alias" "example" {
  alias    = "foo@example.com"
  group_id = "93c429aa-8760-452d-8c3c-e94d96ca102a"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) The alias or ProxyAddress to add to the O365 Group.

* `group_id` - (Required) The GUID of the O365 group to manage. It can also be
  sourced from the `O365_GROUP_ID` environment variable.

## Import

O365 group aliases can be imported using a combination of the group ID and the
alias itself, e.g.

```
$ terraform import mcaf_o365_alias.example 850dbcea-6a4a-46da-adb5-10f1a702f7ab:alias@example.com
```
