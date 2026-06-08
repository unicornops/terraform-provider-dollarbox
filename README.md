# Terraform Provider for DollarBox

Official Terraform provider for DollarBox.

This repository uses the Terraform Plugin Framework. The provider manages
DollarBox resources through the public API.

## Requirements

- Terraform 1.0 or newer
- Go 1.25 or newer
- pre-commit, for local commit hooks
- golangci-lint, for local linting
- GoReleaser and GPG, for local release testing

## Provider Configuration

```hcl
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

resource "dollarbox_container" "web" {
  name  = "web"
  image = "nginx:1.27"

  env = {
    NGINX_PORT = "80"
  }
}

resource "dollarbox_volume" "data" {
  name    = "data"
  size_gb = 10
}

resource "dollarbox_invitation" "admin" {
  email = "admin@example.com"
  role  = "admin"
}

resource "dollarbox_kubectl_credential" "current" {}

data "dollarbox_org" "current" {}
```

`token` can also be provided with the `DOLLARBOX_TOKEN` environment variable.

## Supported Resources and Data Sources

- `dollarbox_container` manages `/api/v1/containers/` resources.
- `dollarbox_volume` manages `/api/v1/volumes/` resources.
- `dollarbox_invitation` manages `/api/v1/invitations/` resources.
- `dollarbox_kubectl_credential` manages `/api/v1/kubectl-credentials/` resources.
- `dollarbox_org` reads `/api/v1/orgs/{slug}/` metadata.

## Development

```shell
make test
make lint
make build
```

Install the local pre-commit and pre-push hooks:

```shell
pre-commit install --install-hooks
```

Run the commit-stage hooks manually:

```shell
pre-commit run --all-files
```

Run the push-stage hooks manually:

```shell
pre-commit run --hook-stage pre-push --all-files
```

## GitHub Actions

- `.github/workflows/build.yml` runs pre-commit, Go formatting, vet, tests,
  golangci-lint, build, and a GoReleaser snapshot on pull requests and pushes
  to `main`.
- `.github/workflows/acceptance.yml` runs `make testacc` only on pushes to
  `main` and manual dispatches. It requires `DOLLARBOX_TOKEN` and
  `DOLLARBOX_ORG` secrets for a dedicated DollarBox test organisation; public
  pull requests never receive these secrets.
- `.github/workflows/release.yml` creates signed provider release artifacts when
  a `v*` tag is pushed.

## Releases

Before the first release, create repository secrets named `GPG_PRIVATE_KEY` and
`PASSPHRASE`. The release workflow imports that key, signs the GoReleaser
checksum file, and uploads Terraform Registry-compatible assets:

- provider zip files
- `terraform-provider-dollarbox_<version>_manifest.json`
- `terraform-provider-dollarbox_<version>_SHA256SUMS`
- `terraform-provider-dollarbox_<version>_SHA256SUMS.sig`

Add the matching public GPG key to the Terraform Registry namespace before
publishing the provider. Release tags must be valid semantic versions prefixed
with `v`, for example `v0.1.0`.
