---
page_title: "dollarbox_org Data Source"
description: |-
  Reads a DollarBox organisation.
---

# dollarbox_org Data Source

Reads a DollarBox organisation through `/api/v1/orgs/{slug}/`.

## Example Usage

```terraform
data "dollarbox_org" "current" {}

data "dollarbox_org" "acme" {
  slug = "acme"
}
```

## Schema

### Optional

- `slug` (String) Organisation slug. Defaults to the provider-level `org` when omitted.

### Read-Only

- `api_enabled` (Boolean) Whether API access is enabled for this organisation.
- `billing_email` (String) Billing email address.
- `billing_mode` (String) Organisation billing mode.
- `created_at` (String) Creation timestamp.
- `kubectl_enabled` (Boolean) Whether kubectl mode is enabled.
- `name` (String) Organisation display name.
- `status` (String) Organisation lifecycle status.
