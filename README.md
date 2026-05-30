# Terraform Provider for DollarBox

Official Terraform provider for DollarBox.

This repository uses the Terraform Plugin Framework. The initial scaffold builds
and validates the provider configuration; resources and data sources will be
added once the DollarBox public API is available.

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
  endpoint = "https://dollarbox.io"
  token    = var.dollarbox_token
  org      = "my-org"
}
```

`token` can also be provided with the `DOLLARBOX_TOKEN` environment variable.

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
