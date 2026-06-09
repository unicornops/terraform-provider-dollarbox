---
page_title: "dollarbox_namespace Data Source"
description: |-
  Reads a DollarBox namespace.
---

# dollarbox_namespace Data Source

Reads a DollarBox namespace through `/api/v1/namespaces/{id}/` or by searching `/api/v1/namespaces/` by slug.

## Example Usage

```terraform
data "dollarbox_namespace" "dev" {
  slug = "dev"
}

data "dollarbox_namespace" "by_id" {
  id = "123"
}
```

## Schema

### Optional

- `id` (String) DollarBox namespace ID.
- `slug` (String) Namespace slug. May be used instead of `id` for lookup.

At least one of `id` or `slug` must be set.

### Read-Only

- `allocated_containers` (Number) Number of container slots allocated to this namespace.
- `allocated_volume_gb` (Number) Volume capacity in GB allocated to this namespace.
- `created_at` (String) Creation timestamp.
- `k8s_namespace` (String) Backing Kubernetes namespace name.
- `status` (String) Current namespace status.
- `updated_at` (String) Last update timestamp.
