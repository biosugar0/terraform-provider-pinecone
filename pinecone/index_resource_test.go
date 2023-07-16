package pinecone

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "pinecone_index" "test" {
	name           = "test"
	dimension      = 1536
	metric         = "dotproduct"
	metadata_config {
		indexed = ["potato"]
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of items
					resource.TestCheckResourceAttr("pinecone_index.test", "name", "test"),
					resource.TestCheckResourceAttr("pinecone_index.test", "id", "test"),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "dotproduct"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pods", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "replicas", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pod_type", "p1.x1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metadata_config.0.indexed.0", "potato"),
				),
			},
			{
				Config: providerConfig + `
resource "pinecone_index" "test" {
	name      = "test"
	dimension = 1536
	metric    = "dotproduct"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pinecone_index.test", "name", "test"),
					resource.TestCheckResourceAttr("pinecone_index.test", "id", "test"),
					resource.TestCheckResourceAttr("pinecone_index.test", "dimension", "1536"),
					resource.TestCheckResourceAttr("pinecone_index.test", "metric", "dotproduct"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pods", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "replicas", "1"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pod_type", "p1.x1"),
					resource.TestCheckNoResourceAttr("pinecone_index.test", "metadata_config"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "pinecone_index.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
			resource "pinecone_index" "test" {
			    name      = "test"
                dimension = 1536
                metric    = "dotproduct"
                replicas  = 2
                pod_type  = "p1.x2"
			}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first order item updated
					resource.TestCheckResourceAttr("pinecone_index.test", "name", "test"),
					resource.TestCheckResourceAttr("pinecone_index.test", "id", "test"),
					resource.TestCheckResourceAttr("pinecone_index.test", "replicas", "2"),
					resource.TestCheckResourceAttr("pinecone_index.test", "pod_type", "p1.x2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
