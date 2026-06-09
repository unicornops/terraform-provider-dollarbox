---
page_title: "dollarbox_members Data Source"
description: |-
  Lists DollarBox organisation members.
---

# dollarbox_members Data Source

Lists DollarBox organisation members through `/api/v1/members/`.

## Example Usage

```terraform
data "dollarbox_members" "current" {}
```

## Schema

### Read-Only

- `members` (List of Object) Members in the selected organisation.

Each member object includes:

- `email` (String) Member email address.
- `id` (String) DollarBox membership ID.
- `joined_at` (String) Membership join timestamp.
- `role` (String) Member role in the organisation.
