terraform {
  required_providers {
    pinecone = {
      source = "registry.terraform.io/biosugar0/pinecone"
    }
  }
}

provider "pinecone" {}

resource "pinecone_index" "test" {
  name      = "test"
  dimension = 1536
  metric    = "dotproduct"
}
