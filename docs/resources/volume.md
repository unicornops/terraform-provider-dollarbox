---
page_title: "dollarbox_volume Resource"
description: |-
  Manages a DollarBox volume.
---

# dollarbox_volume Resource

Manages a DollarBox persistent volume through `/api/v1/volumes/`.

## Example Usage

```terraform
resource "dollarbox_volume" "data" {
  name    = "data"
  size_gb = 10
}
```

## Schema

### Required

- `name` (String) Volume name. Changing this recreates the volume.
- `size_gb` (Number) Volume size in GB. Changing this recreates the volume.

### Read-Only

- `created_at` (String) Creation timestamp.
- `id` (String) DollarBox volume ID.
- `status` (String) Current volume status.
- `storage_class` (String) Kubernetes storage class backing the volume.
- `updated_at` (String) Last update timestamp.

## Import

Import volumes by numeric ID:

```shell
terraform import dollarbox_volume.data 456
```
