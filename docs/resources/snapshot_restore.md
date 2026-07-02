---
page_title: "dollarbox_snapshot_restore Resource"
description: |-
  Restores a DollarBox volume snapshot into a new PVC.
---

# dollarbox_snapshot_restore Resource

Restores a ready snapshot into a new PVC. The API enforces the namespace storage allocation.

Destroying this resource only removes the completed operation from Terraform state. It **does not delete the restored PVC**. Manage or remove that PVC separately.

## Example Usage

```terraform
resource "dollarbox_snapshot_restore" "recovered" {
  namespace_id    = 42
  source_pvc_name = "app-data"
  snapshot_id     = "11111111-2222-3333-4444-555555555555"
  target_pvc_name = "app-data-restored"
}
```

## Schema

### Required

- `namespace_id` (Number) DollarBox namespace ID containing the source PVC.
- `source_pvc_name` (String) PVC containing the source snapshot.
- `snapshot_id` (String) Ready snapshot UUID.
- `target_pvc_name` (String) Name of the new PVC.

All inputs are immutable and changes create a new restore operation.

### Optional

- `timeouts.create` (String) Maximum wait for the target PVC to become bound. Defaults to `30m`.

### Read-Only

- `id` (String) Restore UUID.
- `status` (String) Restore status.
- `error` (String) Latest error, when present.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.
