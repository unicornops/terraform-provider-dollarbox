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
  endpoint = "https://app.dollarbox.dev"
  token    = var.dollarbox_token
  org      = "my-org"
}
