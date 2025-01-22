// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package provider

// import (
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// )

// func TestAccAccelByteMatchPoolDataSource(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Read testing
// 			{
// 				Config: testAccAccelByteMatchPoolDataSourceConfig,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("data.accelbyte_match_pool.test_pool", "id", "namespace1/pool01"),
// 				),
// 			},
// 		},
// 	})
// }

// const testAccAccelByteMatchPoolDataSourceConfig = `
// provider "accelbyte" {
// 	base_url = "https://localhost"
// 	iam_client_id = "abcd"
// 	iam_client_secret = "efgh"
// 	admin_username = "user@example.com"
// 	admin_password = "pass"
// }

// data "accelbyte_match_pool" "test_pool" {
//   namespace = "namespace1"
//   name = "pool01"
// }
// `
