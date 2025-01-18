terraform {
  required_providers {
    accelbyte = {
      source = "WalkerLabs/accelbyte"
    }
  }
}

provider "accelbyte" {}

data "accelbyte_example" "example" {
  configurable_attribute = "some-value"
}
