---
page_title: "dollarbox_kubectl_credential Resource"
description: |-
  Manages a DollarBox kubectl credential.
---

# dollarbox_kubectl_credential Resource

Issues and revokes a DollarBox kubectl credential for the authenticated user and selected organisation through `/api/v1/kubectl-credentials/`.

## Example Usage

```terraform
resource "dollarbox_kubectl_credential" "current" {}
```

## Schema

### Read-Only

- `created_at` (String) Creation timestamp.
- `id` (String) DollarBox kubectl credential ID.
- `kubeconfig` (String, Sensitive) Kubeconfig returned when the credential is issued. The API only returns this value on create.
- `last_downloaded_at` (String) Last kubeconfig download timestamp.
- `org` (String) Organisation slug for the credential.
- `rotated_at` (String) Last rotation timestamp.
- `sa_name` (String) Kubernetes service account name.

## Import

Import kubectl credentials by numeric ID. Imported resources do not receive `kubeconfig`, because the API only returns that value when issuing a credential:

```shell
terraform import dollarbox_kubectl_credential.current 321
```
