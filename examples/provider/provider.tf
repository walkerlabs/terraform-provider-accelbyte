terraform {
  required_providers {
    accelbyte = {
      source = "walkerlabs/accelbyte"
    }
  }
}

provider "accelbyte" {
  base_url          = "https://<something>.accelbyte.io" # or set via ACCELBYTE_BASE_URL
  iam_client_id     = "<typically a hex string>"         # or set via ACCELBYTE_IAM_CLIENT_ID
  iam_client_secret = "<...>"                            # or set via ACCELBYTE_IAM_CLIENT_SECRET
  admin_username    = "<typically an email address>"     # or set via ACCELBYTE_ADMIN_USERNAME
  admin_password    = "<...>"                            # or set via ACCELBYTE_ADMIN_PASSWORD
}
