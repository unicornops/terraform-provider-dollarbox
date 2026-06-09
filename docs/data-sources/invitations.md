---
page_title: "dollarbox_invitations Data Source"
description: |-
  Lists DollarBox invitations.
---

# dollarbox_invitations Data Source

Lists DollarBox organisation invitations through `/api/v1/invitations/`.

## Example Usage

```terraform
data "dollarbox_invitations" "current" {}
```

## Schema

### Read-Only

- `invitations` (List of Object) Invitations in the selected organisation.

Each invitation object includes `id`, `email`, `role`, `accepted`, `expires_at`, and `created_at`.
