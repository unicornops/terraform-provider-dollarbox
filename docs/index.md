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
  endpoint = "https://app.dollarbox.dev"
  token    = var.dollarbox_token
  org      = "my-org"
}

resource "dollarbox_org" "current" {
  slug          = "my-org"
  name          = "My Org"
  billing_email = "billing@example.com"
}

resource "dollarbox_container" "web" {
  name  = "web"
  image = "nginx:1.27"
}

resource "dollarbox_volume" "data" {
  name    = "data"
  size_gb = 10
}

resource "dollarbox_invitation" "admin" {
  email = "admin@example.com"
  role  = "admin"
}

resource "dollarbox_member" "admin" {
  email = "admin@example.com"
  role  = "admin"
}

resource "dollarbox_namespace" "dev" {
  slug                 = "dev"
  allocated_containers = 2
  allocated_volume_gb  = 10
}

resource "dollarbox_kubectl_credential" "current" {}

data "dollarbox_org" "current" {}
data "dollarbox_containers" "current" {}
data "dollarbox_volumes" "current" {}
data "dollarbox_invitations" "current" {}
data "dollarbox_kubectl_credentials" "current" {}
data "dollarbox_members" "current" {}
data "dollarbox_namespaces" "current" {}
data "dollarbox_orgs" "current" {}
```

## Schema

### Optional

- `endpoint` (String) DollarBox API endpoint. Defaults to `https://app.dollarbox.dev`.
- `org` (String) Optional DollarBox organisation slug used as the default scope.
- `token` (String, Sensitive) DollarBox API token. May also be set with `DOLLARBOX_TOKEN`.
