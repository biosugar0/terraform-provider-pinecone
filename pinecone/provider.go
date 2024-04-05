package pinecone

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &pineconeProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(cli PineconeClientInterface, version string) provider.Provider {
	return &pineconeProvider{
		client:  cli,
		version: version,
	}
}

// pineconeProvider is the provider implementation.
type pineconeProvider struct {
	client  PineconeClientInterface
	version string
}

// hashicupsProviderModel maps provider schema data to a Go type.
type pineconeProviderModel struct {
	Environment types.String `tfsdk:"environment"`
	ApiKey      types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *pineconeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pinecone"
	resp.Version = "v0.0.6"
}

// Schema defines the provider-level schema for configuration data.
func (p *pineconeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Pinecone vector database. https://www.pinecone.io/ ",
		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				Description: "The Pinecone environment to use.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The Pinecone API key to use.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a pinecone API client for data sources and resources.
func (p *pineconeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Pinecone client")

	// Retrieve provider data from configuration
	var config pineconeProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown pinecone API Key",
			"The provider cannot create the pinecone API client as there is an unknown configuration value for the Pinecone API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the pinecone_HOST environment variable.",
		)
	}

	if config.Environment.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("environment"),
			"Unknown pinecone environment",
			"The provider cannot create the pinecone API client as there is an unknown configuration value for the Pinecone environment. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the pinecone_ENVIRONMENT environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("PINECONE_API_KEY")
	environment := os.Getenv("PINECONE_ENVIRONMENT")

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	if !config.Environment.IsNull() {
		environment = config.Environment.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if environment == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("environment"),
			"Missing Pinecone API environment",
			"The provider cannot create the Pinecone API client as there is a missing or empty value for the Pinecone API environment. "+
				"Set the environment value in the configuration or use the PINECONE_ENVIRONMENT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Pinecone API Key",
			"The provider cannot create the Pinecone API client as there is a missing or empty value for the Pinecone API Key. "+
				"Set the api_key value in the configuration or use the PINECONE_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "pinecone_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "pinecone_api_key")
	ctx = tflog.SetField(ctx, "pinecone_environment", environment)

	tflog.Debug(ctx, "Creating Pinecone client")

	// Create a Pinecone API client using the configuration values.
	client, err := NewClient(apiKey, environment)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Pinecone API Client",
			"An unexpected error occurred when creating the Pinecone API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Pinecone Client Error: "+err.Error(),
		)
		return
	}

	// Make the Pinecone API client available to data sources and resources.
	if p.client != nil {
		resp.DataSourceData = p.client
		resp.ResourceData = p.client
		tflog.Info(ctx, "Configured Mock Pinecone client", map[string]any{"success": true})
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "Configured Pinecone client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *pineconeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIndexDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *pineconeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIndexResource,
	}
}
