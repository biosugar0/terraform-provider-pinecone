package main

import (
	"context"
	"fmt"
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
			// this provider has been archived. need notice.
			fmt.Println("This provider has been archived. Please use the other provider.")
			return pinecone.New(nil)
		},
		providerserver.ServeOpts{
			Address: "registry.terraform.io/biosugar0/pinecone",
		})
}
