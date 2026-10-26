---
name: Bug Report
about: Report a bug in the OpenMetadata Terraform provider
title: "bug: "
labels: bug
assignees: ""
---

## Bug Description

<!-- Clear, concise description of the bug -->

## Steps to Reproduce

1. Terraform config:
```hcl
resource "openmetadata_..." "example" {
  # ...
}
```
2. Run `terraform apply`
3. See error

## Expected Behavior

<!-- What should happen? -->

## Actual Behavior

<!-- What actually happened? Include error output -->

```
Error: ...
```

## Environment

- Provider version: `v0.x.x`
- Terraform version: `terraform version`
- Go version (if building from source): `go version`
- OpenMetadata version:
- OpenMetadata authentication mode (token or username/password):
- OS:

## Additional Context

<!-- Logs, TF_LOG=DEBUG output with tokens, passwords and hostnames removed -->
