terraform {
  required_providers {
    dollarbox = {
      source = "unicornops/dollarbox"
    }
  }
}

variable "dollarbox_token" {
  type      = string
  nullable  = false
  sensitive = true
}

provider "dollarbox" {
  endpoint = "https://dollarbox.io"
  token    = var.dollarbox_token
  org      = "my-org"
}
