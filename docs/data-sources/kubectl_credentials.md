---
page_title: "dollarbox_kubectl_credentials Data Source"
description: |-
  Lists DollarBox kubectl credentials.
---

# dollarbox_kubectl_credentials Data Source

Lists DollarBox kubectl credentials through `/api/v1/kubectl-credentials/`.

## Example Usage

```terraform
data "dollarbox_kubectl_credentials" "current" {}
```

## Schema

### Read-Only

- `kubectl_credentials` (List of Object) Kubectl credentials for the authenticated user.

Each credential object includes `id`, `org`, `sa_name`, `kubeconfig`, `created_at`, `rotated_at`, and `last_downloaded_at`. The API only returns `kubeconfig` when issuing a new credential, so list results normally have it unset.
