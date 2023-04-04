terraform {
  required_providers {
    pinecone = {
      source = "registry.terraform.io/biosugar0/pinecone"
    }
  }
}

provider "pinecone" {}

data "pinecone_index" "example" {}
