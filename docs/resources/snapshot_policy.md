---
page_title: "dollarbox_snapshot_policy Resource"
description: |-
  Manages daily snapshot protection for a DollarBox PVC.
---

# dollarbox_snapshot_policy Resource

Enables daily snapshot protection for a bound Longhorn PVC in a child namespace.

Snapshot protection costs **€0.10 per protected source GB per month**. Billing is based on the source PVC capacity, not physical snapshot bytes. Charges increase before activation completes and decrease only after all retained CSI snapshots have been deleted. Snapshots are same-cluster crash-consistent recovery points, not off-site backups.

## Example Usage

```terraform
resource "dollarbox_snapshot_policy" "data" {
  namespace_id   = 42
  pvc_name       = "app-data"
  retention_days = 7
}
```

## Schema

### Required

- `namespace_id` (Number) DollarBox namespace ID containing the PVC. Changing it replaces the policy.
- `pvc_name` (String) Kubernetes PVC name to protect. Changing it replaces the policy.

### Optional

- `retention_days` (Number) Daily snapshots to retain, from 1 through 7. Defaults to `7`.
- `timeouts.create` (String) Maximum activation wait. Defaults to `30m`.
- `timeouts.update` (String) Maximum update wait. Defaults to `30m`.
- `timeouts.delete` (String) Maximum retirement wait. Defaults to `30m`.

### Read-Only

- `id` (String) Compound ID in `namespace_id/pvc_name` form.
- `protected_gb` (Number) Source PVC capacity protected in GB.
- `billed_gb` (Number) Protected capacity currently billed in GB.
- `monthly_cost_cents` (Number) Current monthly policy cost in euro cents.
- `status` (String) Policy status.
- `next_snapshot_at` (String) Scheduled next snapshot time, when available.
- `last_snapshot_at` (String) Most recent snapshot time, when available.
- `error` (String) Latest error, when present.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

Destroy waits for status `disabled`, because billing continues while retained snapshots are being deleted.

## Import

Import a policy with its namespace ID and PVC name:

```shell
terraform import dollarbox_snapshot_policy.data 42/app-data
```
