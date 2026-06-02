---
page_title: "dollarbox_invitation Resource"
description: |-
  Manages a DollarBox organisation invitation.
---

# dollarbox_invitation Resource

Manages a DollarBox organisation invitation through `/api/v1/invitations/`.

## Example Usage

```terraform
resource "dollarbox_invitation" "admin" {
  email = "admin@example.com"
  role  = "admin"
}
```

## Schema

### Required

- `email` (String) Email address to invite. Changing this recreates the invitation.

### Optional

- `role` (String) Role to grant when the invitation is accepted. Defaults to `member`. The API rejects `owner` invitations. Changing this recreates the invitation.

### Read-Only

- `accepted` (Boolean) Whether the invitation has been accepted.
- `created_at` (String) Creation timestamp.
- `expires_at` (String) Invitation expiry timestamp.
- `id` (String) DollarBox invitation ID.

## Import

Import invitations by numeric ID:

```shell
terraform import dollarbox_invitation.admin 789
```
