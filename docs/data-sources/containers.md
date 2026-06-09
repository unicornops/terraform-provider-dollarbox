---
page_title: "dollarbox_containers Data Source"
description: |-
  Lists DollarBox containers.
---

# dollarbox_containers Data Source

Lists DollarBox containers through `/api/v1/containers/`.

## Example Usage

```terraform
data "dollarbox_containers" "current" {}
```

## Schema

### Read-Only

- `containers` (List of Object) Containers in the selected organisation.

Each container object includes `id`, `name`, `image`, `env`, `command`, `status`, `ipv6_address`, `created_at`, and `updated_at`.
