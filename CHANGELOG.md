## 0.4.3 (2024-10-04)

- Support newer "AWS Control Tower Account Factory" versions

## 0.4.2 (2022-11-02)

- Add `mcaf_aws_all_organizational_units` data resource
- Fix `mcaf_aws_account` so it can move accounts to new organizational units
- Update `mcaf_aws_account` docs
- Update `mcaf_aws_account` to provision accounts using updated Service Catalog fields
- Update module versions and require minimum Go 1.18 to build plugin

## 0.4.1 (2022-03-08)

- Add missing docs for the new CodeBuild resource.

## 0.4.0 (2022-03-08)

- Add a new resource to trigger CodeBuild pipelines.

## 0.3.1 (2022-01-27)

- Build binaries using Go 1.17.

## 0.3.0 (2022-01-27)

- Remove all o365 code as it's not needed anymore.

## 0.2.0 (2020-11-30)

- Switch to using v2 of the ExoAPI; v1 users will need to stick to 0.1.x versions of this provider.

## 0.1.2 (2020-11-16)

- Make sure all `mcaf_aws_account` CRUD actions are behind a mutex (only one action can be executed at a time).

## 0.1.1 (2020-10-23)

- Fix bug in `mcaf_aws_account` where failed AWS account provisioning attempt via Service Catalog did not result in a tainted resource.

## 0.1.0 (2020-09-28)

Initial release.
