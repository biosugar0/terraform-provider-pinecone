package pinecone

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &indexDataSource{}
	_ datasource.DataSourceWithConfigure = &indexDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewIndexDataSource() datasource.DataSource {
	return &indexDataSource{}
}

// coffeesDataSource is the data source implementation.
type indexDataSource struct {
	client PineconeClientInterface
}

type indexDataSourceModel struct {
	ID             types.String    `tfsdk:"id"`
	Name           types.String    `tfsdk:"name"`
	Metric         types.String    `tfsdk:"metric"`
	Dimension      types.Int64     `tfsdk:"dimension"`
	Replicas       types.Int64     `tfsdk:"replicas"`
	Shards         types.Int64     `tfsdk:"shards"`
	Pods           types.Int64     `tfsdk:"pods"`
	PodType        types.String    `tfsdk:"pod_type"`
	MetadataConfig *metadataConfig `tfsdk:"metadata_config"`
	Status         *indexStatus    `tfsdk:"status"`
}

type indexStatus struct {
	Host  types.String `tfsdk:"host"`
	Port  types.Int64  `tfsdk:"port"`
	State types.String `tfsdk:"state"`
	Ready types.Bool   `tfsdk:"ready"`
}

type metadataConfig struct {
	Indexed types.List `tfsdk:"indexed"`
}

// Metadata returns the data source type name.
func (d *indexDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index"
}

// Schema defines the schema for the data source.
func (d *indexDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get information about an index.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the index.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the index.",
				Required:    true,
			},
			"metric": schema.StringAttribute{
				Description: "The metric of the index.",
				Computed:    true,
			},
			"dimension": schema.Int64Attribute{
				Description: "The dimension of the index.",
				Computed:    true,
			},
			"replicas": schema.Int64Attribute{
				Description: "The replicas of the index.",
				Computed:    true,
			},
			"shards": schema.Int64Attribute{
				Description: "The shards of the index.",
				Computed:    true,
			},
			"pods": schema.Int64Attribute{
				Description: "The pods of the index.",
				Computed:    true,
			},
			"pod_type": schema.StringAttribute{
				Description: "The pod type of the index.",
				Computed:    true,
			},
			"metadata_config": schema.SingleNestedAttribute{
				Description: "The metadata config of the index.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"indexed": schema.ListAttribute{
						Description: "The indexed of the index.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"status": schema.SingleNestedAttribute{
				Description: "The status of the index.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Description: "The host of the index.",
						Computed:    true,
					},
					"port": schema.Int64Attribute{
						Description: "The port of the index.",
						Computed:    true,
					},
					"state": schema.StringAttribute{
						Description: "The state of the index.",
						Computed:    true,
					},
					"ready": schema.BoolAttribute{
						Description: "The ready state of the index.",
						Computed:    true,
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.

func (d *indexDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data indexDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	index, err := d.client.DescribeIndex(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error DescribeIndex", err.Error())
		return
	}

	if index == nil {
		// Set an empty state if index is not found
		emptyState := indexDataSourceModel{
			ID: types.StringValue(data.Name.ValueString()),
		}
		resp.State.Set(ctx, &emptyState)
		return
	}

	// Define MetadataConfig
	var metaConfig *metadataConfig
	if index.Database.MetadataConfig != nil {
		metaConfig = &metadataConfig{}
		// Use ListValueFrom to create List from a slice of strings
		listVal, diags := types.ListValueFrom(ctx, types.StringType, index.Database.MetadataConfig.Indexed)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		metaConfig.Indexed = listVal
	}

	state := indexDataSourceModel{
		ID:             types.StringValue(data.Name.ValueString()), // Set a unique value for the ID field
		Name:           types.StringValue(index.Database.Name),
		Metric:         types.StringValue(index.Database.Metric.String()),
		Dimension:      types.Int64Value(int64(index.Database.Dimension)),
		Replicas:       types.Int64Value(int64(index.Database.Replicas)),
		Shards:         types.Int64Value(int64(index.Database.Shards)),
		Pods:           types.Int64Value(int64(index.Database.Pods)),
		PodType:        types.StringValue(index.Database.PodType.String()),
		MetadataConfig: metaConfig,
		Status: &indexStatus{
			Host:  types.StringValue(index.Status.Host),
			Port:  types.Int64Value(int64(index.Status.Port)),
			State: types.StringValue(index.Status.State),
			Ready: types.BoolValue(index.Status.Ready),
		},
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *indexDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(PineconeClientInterface)

	if !ok {
		resp.Diagnostics.AddError("Error Configure", "Invalid provider data")
		return
	}

	d.client = client
}
