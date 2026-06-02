---
page_title: "dollarbox_container Resource"
description: |-
  Manages a DollarBox container.
---

# dollarbox_container Resource

Manages a DollarBox container through `/api/v1/containers/`.

## Example Usage

```terraform
resource "dollarbox_container" "web" {
  name  = "web"
  image = "nginx:1.27"

  env = {
    NGINX_PORT = "80"
  }

  command = ["nginx", "-g", "daemon off;"]
}
```

## Schema

### Required

- `image` (String) OCI image reference to run.
- `name` (String) Container name. Changing this recreates the container.

### Optional

- `command` (List of String) Container command override.
- `env` (Map of String, Sensitive) Environment variables for the container.

### Read-Only

- `created_at` (String) Creation timestamp.
- `id` (String) DollarBox container ID.
- `ipv6_address` (String) Assigned IPv6 LoadBalancer address.
- `status` (String) Current container status.
- `updated_at` (String) Last update timestamp.

## Import

Import containers by numeric ID:

```shell
terraform import dollarbox_container.web 123
```
