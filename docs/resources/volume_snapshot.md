---
page_title: "dollarbox_volume_snapshot Resource"
description: |-
  Manages a manual DollarBox volume snapshot.
---

# dollarbox_volume_snapshot Resource

Creates the single manual snapshot permitted for a protected Longhorn PVC. Use the snapshot data sources to reference daily snapshots owned by the managed policy.

## Example Usage

```terraform
resource "dollarbox_volume_snapshot" "before_upgrade" {
  namespace_id = 42
  pvc_name     = "app-data"
  name         = "before-upgrade"
  labels = {
    environment = "production"
  }
}
```

## Schema

### Required

- `namespace_id` (Number) DollarBox namespace ID containing the PVC.
- `pvc_name` (String) Source PVC name.

### Optional

- `name` (String) Display name, up to 100 characters.
- `labels` (Map of String) Up to 10 Kubernetes labels accepted by the API.
- `timeouts.create` (String) Maximum readiness wait. Defaults to `30m`.
- `timeouts.delete` (String) Maximum deletion wait. Defaults to `30m`.

All input changes replace the immutable snapshot.

### Read-Only

- `id` (String) Snapshot UUID.
- `kind` (String) Snapshot kind.
- `status` (String) Snapshot status.
- `restore_size_bytes` (Number) Required restore size reported by CSI, when available.
- `ready_at` (String) Readiness timestamp, when available.
- `error` (String) Latest error, when present.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

## Import

Import a manual snapshot using the namespace, PVC, and snapshot UUID:

```shell
terraform import dollarbox_volume_snapshot.before_upgrade 42/app-data/11111111-2222-3333-4444-555555555555
```
