terraform {
  required_providers {
    pinecone = {
      source = "registry.terraform.io/biosugar0/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_index" "test" {
  name = "test"
}

output "test_indices" {
  value = data.pinecone_index.test
}
