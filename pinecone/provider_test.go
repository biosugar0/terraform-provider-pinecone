package pinecone

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the HashiCups client is properly configured.
	// It is also possible to use the HASHICUPS_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `
provider "pinecone" {
    environment = "test"
    api_key = "test_api_key"
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"pinecone": providerserver.NewProtocol6WithError(
			func() provider.Provider {
				cli, err := NewMockClient(
					WithAPIKey("test_api_key"),
					WithEnvironment("test"),
				)

				if err != nil {
					panic(err)
				}
				return New(cli)
			}()),
	}
)
