---
page_title: "DollarBox Provider"
description: |-
  The DollarBox provider manages DollarBox resources through the public API.
---

# DollarBox Provider

The DollarBox provider manages DollarBox resources through the public API.

## Example Usage

```terraform
terraform {
  required_providers {
    dollarbox = {
      source = "unicornops/dollarbox"
    }
  }
}

provider "dollarbox" {
  endpoint = "https://dollarbox.io"
  token    = var.dollarbox_token
  org      = "my-org"
}
```

## Schema

### Optional

- `endpoint` (String) DollarBox API endpoint. Defaults to `https://dollarbox.io`.
- `org` (String) Optional DollarBox organisation slug used as the default scope.
- `token` (String, Sensitive) DollarBox API token. May also be set with `DOLLARBOX_TOKEN`.
