---
page_title: "dollarbox_namespaces Data Source"
description: |-
  Lists DollarBox namespaces.
---

# dollarbox_namespaces Data Source

Lists DollarBox namespaces through `/api/v1/namespaces/`.

## Example Usage

```terraform
data "dollarbox_namespaces" "current" {}
```

## Schema

### Read-Only

- `namespaces` (List of Object) Namespaces in the selected organisation.

Each namespace object includes:

- `allocated_containers` (Number) Number of container slots allocated to this namespace.
- `allocated_volume_gb` (Number) Volume capacity in GB allocated to this namespace.
- `created_at` (String) Creation timestamp.
- `id` (String) DollarBox namespace ID.
- `k8s_namespace` (String) Backing Kubernetes namespace name.
- `slug` (String) Namespace slug.
- `status` (String) Current namespace status.
- `updated_at` (String) Last update timestamp.
