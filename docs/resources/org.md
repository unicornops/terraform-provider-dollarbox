---
page_title: "dollarbox_org Resource"
description: |-
  Manages DollarBox organisation settings.
---

# dollarbox_org Resource

Manages settings for an existing DollarBox organisation through `/api/v1/orgs/{slug}/`.

The DollarBox API does not create or delete orgs through Terraform. Destroying this resource removes Terraform management only; it does not delete the org.

## Example Usage

```terraform
resource "dollarbox_org" "current" {
  slug          = "my-org"
  name          = "My Org"
  billing_email = "billing@example.com"
}
```

## Schema

### Required

- `billing_email` (String) Billing email address.
- `name` (String) Organisation display name.
- `slug` (String) Organisation slug. Changing this adopts a different org and removes the old one from Terraform state during replacement.

### Read-Only

- `api_enabled` (Boolean) Whether API access is enabled for this organisation.
- `billing_mode` (String) Organisation billing mode.
- `created_at` (String) Creation timestamp.
- `kubectl_enabled` (Boolean) Whether kubectl mode is enabled for this organisation.
- `status` (String) Organisation lifecycle status.

## Import

Import org settings by slug:

```shell
terraform import dollarbox_org.current my-org
```
