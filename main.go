package main

import (
	"context"
	"terraform-provider-pinecone/pinecone"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

const (
	version        = "v0.0.6"
	warningMessage = ":This provider has been archived. Please use the other provider."
)

// Provider documentation generation.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name pinecone
func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/biosugar0/pinecone",
	}

	err := providerserver.Serve(context.Background(), func() provider.Provider {
		return pinecone.New(nil, version+warningMessage)
	}, opts)

	if err != nil {
		panic(err)
	}
}
