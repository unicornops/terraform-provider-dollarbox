---
page_title: "dollarbox_volumes Data Source"
description: |-
  Lists DollarBox volumes.
---

# dollarbox_volumes Data Source

Lists DollarBox volumes through `/api/v1/volumes/`.

## Example Usage

```terraform
data "dollarbox_volumes" "current" {}
```

## Schema

### Read-Only

- `volumes` (List of Object) Volumes in the selected organisation.

Each volume object includes `id`, `name`, `size_gb`, `status`, `storage_class`, `created_at`, and `updated_at`.
