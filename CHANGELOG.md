## 0.1.2 (2020-11-16)

- Make sure all `mcaf_aws_account` CRUD actions are behind a mutex (only one action can be executed at a time).

## 0.1.1 (2020-10-23)

- Fix bug in `mcaf_aws_account` where failed AWS account provisioning attempt via Service Catalog did not result in a tainted resource.

## 0.1.0 (2020-09-28)

Initial release.
