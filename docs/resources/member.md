---
page_title: "dollarbox_member Resource"
description: |-
  Manages a DollarBox organisation member.
---

# dollarbox_member Resource

Manages an existing DollarBox organisation member through `/api/v1/members/{id}/`.

The DollarBox API does not create members directly through `/api/v1/members/`. Use `dollarbox_invitation` to invite a user, wait for acceptance, then use this resource to adopt the accepted member, manage their role, and remove the membership on destroy.

## Example Usage

```terraform
resource "dollarbox_member" "admin" {
  email = "admin@example.com"
  role  = "admin"
}
```

## Schema

### Required

- `email` (String) Email address of an existing organisation member to adopt. Changing this adopts a different member and removes the old membership during replacement.
- `role` (String) Member role in the organisation.

### Read-Only

- `id` (String) DollarBox membership ID.
- `joined_at` (String) Membership join timestamp.

## Import

Import members by numeric ID:

```shell
terraform import dollarbox_member.admin 123
```
