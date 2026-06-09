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

resource "dollarbox_org" "current" {
  slug          = "my-org"
  name          = "My Org"
  billing_email = "billing@example.com"
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

`token` can also be provided with the `DOLLARBOX_TOKEN` environment variable.

## Supported Resources and Data Sources

Resources:

- `dollarbox_container` manages `/api/v1/containers/` resources.
- `dollarbox_volume` manages `/api/v1/volumes/` resources.
- `dollarbox_invitation` manages `/api/v1/invitations/` resources.
- `dollarbox_kubectl_credential` manages `/api/v1/kubectl-credentials/` resources.
- `dollarbox_member` manages accepted organisation members through `/api/v1/members/{id}/`.
- `dollarbox_namespace` manages `/api/v1/namespaces/` resources.
- `dollarbox_org` manages existing org settings through `/api/v1/orgs/{slug}/`.

Data sources:

- `dollarbox_containers` lists `/api/v1/containers/` metadata.
- `dollarbox_volumes` lists `/api/v1/volumes/` metadata.
- `dollarbox_invitations` lists `/api/v1/invitations/` metadata.
- `dollarbox_kubectl_credentials` lists `/api/v1/kubectl-credentials/` metadata.
- `dollarbox_member` reads one member by ID or email.
- `dollarbox_members` lists `/api/v1/members/` metadata.
- `dollarbox_namespace` reads one namespace by ID or slug.
- `dollarbox_namespaces` lists `/api/v1/namespaces/` metadata.
- `dollarbox_org` reads `/api/v1/orgs/{slug}/` metadata.
- `dollarbox_orgs` lists `/api/v1/orgs/` metadata.

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
- `.github/workflows/release.yml` runs release-please on pushes to `main`. When
  a release PR is merged, it creates the `v*` tag and GitHub release, then runs
  GoReleaser to attach signed Terraform Registry-compatible provider artifacts.

## Releases

Before the first release, create repository secrets named `GPG_PRIVATE_KEY` and
`PASSPHRASE`. The release workflow imports that key, signs the GoReleaser
checksum file, and uploads Terraform Registry-compatible assets. Optionally add a
`RELEASE_PLEASE_TOKEN` secret from a GitHub App or fine-grained PAT if release PRs
must trigger required pull request checks:

- provider zip files
- `terraform-provider-dollarbox_<version>_manifest.json`
- `terraform-provider-dollarbox_<version>_SHA256SUMS`
- `terraform-provider-dollarbox_<version>_SHA256SUMS.sig`

Add the matching public GPG key to the Terraform Registry namespace before
publishing the provider.

Releases are managed by release-please. Merge normal feature and fix PRs to
`main` using Conventional Commit titles such as `feat(provider): add namespace
support` or `fix(volume): ignore volatile import fields`. Release-please keeps a
release PR open with generated `CHANGELOG.md` entries. Merging that release PR
creates a draft GitHub release and a semantic version tag prefixed with `v`;
GoReleaser then uploads the signed provider artifacts and publishes the release.

The first release is seeded from `0.0.0` and will publish `v0.1.0` from the
existing feature history once the first release-please PR is merged.
