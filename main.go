package main

import (
	"context"
	"terraform-provider-pinecone/pinecone"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name pinecone
func main() {
	_ = providerserver.Serve(context.Background(),
		func() provider.Provider {
			return pinecone.New(nil)
		},
		providerserver.ServeOpts{
			Address: "registry.terraform.io/biosugar0/pinecone",
		})
}
