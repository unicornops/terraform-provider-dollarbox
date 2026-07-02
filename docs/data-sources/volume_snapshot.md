---
page_title: "dollarbox_volume_snapshot Data Source"
description: |-
  Reads a DollarBox volume snapshot.
---

# dollarbox_volume_snapshot Data Source

Reads a manual or managed daily snapshot without taking Terraform ownership of it.

## Example Usage

```terraform
data "dollarbox_volume_snapshot" "daily" {
  namespace_id = 42
  pvc_name     = "app-data"
  id           = "11111111-2222-3333-4444-555555555555"
}
```

## Schema

### Required

- `namespace_id` (Number) DollarBox namespace ID containing the PVC.
- `pvc_name` (String) Source PVC name.
- `id` (String) Snapshot UUID.

### Read-Only

- `name` (String) Snapshot display name.
- `labels` (Map of String) Snapshot labels.
- `kind` (String) Snapshot kind.
- `status` (String) Snapshot status.
- `restore_size_bytes` (Number) Required restore size, when available.
- `ready_at` (String) Readiness timestamp, when available.
- `error` (String) Latest error, when present.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.
