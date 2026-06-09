---
page_title: "dollarbox_member Data Source"
description: |-
  Reads a DollarBox organisation member.
---

# dollarbox_member Data Source

Reads a DollarBox organisation member through `/api/v1/members/{id}/` or by searching `/api/v1/members/` by email.

## Example Usage

```terraform
data "dollarbox_member" "admin" {
  email = "admin@example.com"
}

data "dollarbox_member" "by_id" {
  id = "123"
}
```

## Schema

### Optional

- `email` (String) Member email address. May be used instead of `id` for lookup.
- `id` (String) DollarBox membership ID.

At least one of `id` or `email` must be set.

### Read-Only

- `joined_at` (String) Membership join timestamp.
- `role` (String) Member role in the organisation.
