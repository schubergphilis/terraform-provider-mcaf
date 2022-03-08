---
layout: "mcaf"
page_title: "MCAF: mcaf_aws_codebuild_trigger"
sidebar_current: "docs-mcaf-resource-aws-codebuild-trigger"
description: |-
  Triggers a CodeBuild pipeline using the configured AWS provider.
---

# mcaf_aws_codebuild_trigger

Triggers a CodeBuild pipeline using the configured AWS provider.

## Example Usage

```hcl
resource "mcaf_aws_codebuild_trigger" "example" {
  project    = "foo"
  release_id = "e58df79"
  version    = "v0.1.0"
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required) The name of the AWS CodeBuild build project.

* `release_id` - (Required) Release ID, used to trigger a release with the same version.

* `version` - (Required) The source version of the build input to be built.
