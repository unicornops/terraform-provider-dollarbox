---
page_title: "dollarbox_orgs Data Source"
description: |-
  Lists DollarBox organisations.
---

# dollarbox_orgs Data Source

Lists DollarBox organisations the token's user belongs to through `/api/v1/orgs/`.

## Example Usage

```terraform
data "dollarbox_orgs" "current" {}
```

## Schema

### Read-Only

- `orgs` (List of Object) Organisations visible to the token.

Each org object includes:

- `api_enabled` (Boolean) Whether API access is enabled for this organisation.
- `billing_email` (String) Billing email address.
- `billing_mode` (String) Organisation billing mode.
- `created_at` (String) Creation timestamp.
- `kubectl_enabled` (Boolean) Whether kubectl mode is enabled for this organisation.
- `name` (String) Organisation display name.
- `slug` (String) Organisation slug.
- `status` (String) Organisation lifecycle status.
