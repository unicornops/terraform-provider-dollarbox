---
page_title: "dollarbox_namespace Resource"
description: |-
  Manages a DollarBox namespace.
---

# dollarbox_namespace Resource

Manages a DollarBox child namespace through `/api/v1/namespaces/`.

## Example Usage

```terraform
resource "dollarbox_namespace" "dev" {
  slug                 = "dev"
  allocated_containers = 2
  allocated_volume_gb  = 10
}
```

## Schema

### Required

- `allocated_containers` (Number) Number of container slots allocated to this namespace.
- `allocated_volume_gb` (Number) Volume capacity in GB allocated to this namespace.
- `slug` (String) Namespace slug. Changing this recreates the namespace.

### Read-Only

- `created_at` (String) Creation timestamp.
- `id` (String) DollarBox namespace ID.
- `k8s_namespace` (String) Backing Kubernetes namespace name.
- `status` (String) Current namespace status.
- `updated_at` (String) Last update timestamp.

## Import

Import namespaces by numeric ID:

```shell
terraform import dollarbox_namespace.dev 123
```
