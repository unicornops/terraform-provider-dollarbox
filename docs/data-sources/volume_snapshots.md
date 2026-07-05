---
page_title: "dollarbox_volume_snapshots Data Source"
description: |-
  Lists DollarBox volume snapshots.
---

# dollarbox_volume_snapshots Data Source

Lists all manual and managed daily snapshots for a PVC. The provider follows every API cursor page.

## Example Usage

```terraform
data "dollarbox_volume_snapshots" "data" {
  namespace_id = 42
  pvc_name     = "app-data"
}
```

## Schema

### Required

- `namespace_id` (Number) DollarBox namespace ID containing the PVC.
- `pvc_name` (String) Source PVC name.

### Read-Only

- `snapshots` (List of Object) Snapshot records. Each record contains `namespace_id`, `pvc_name`, `id`, `name`, `labels`, `kind`, `status`, `restore_size_bytes`, `ready_at`, `error`, `created_at`, and `updated_at`.
