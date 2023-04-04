package pinecone

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &indexResource{}
	_ resource.ResourceWithConfigure   = &indexResource{}
	_ resource.ResourceWithImportState = &indexResource{}
)

// NewIndexResource is a helper function to simplify the provider implementation.
func NewIndexResource() resource.Resource {
	return &indexResource{}
}

// indexResource is the resource implementation.
type indexResource struct {
	client PineconeClientInterface
}

type indexResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Dimension   types.Int64  `tfsdk:"dimension"`
	Metric      types.String `tfsdk:"metric"`
	Pods        types.Int64  `tfsdk:"pods"`
	Replicas    types.Int64  `tfsdk:"replicas"`
	PodType     types.String `tfsdk:"pod_type"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *indexResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index"
}

// Schema defines the schema for the resource.
func (r *indexResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dimension": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"metric": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("cosine"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pods": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
			},
			"pod_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("p1.x1"),
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *indexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan indexResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	metric, err := NewMetric(plan.Metric.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating index",
			"Could not create index, unexpected error: "+err.Error(),
		)
		return
	}
	podType, err := NewPodType(plan.PodType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating index",
			"Could not create index, unexpected error: "+err.Error(),
		)
		return
	}

	item := CreateIndexRequest{
		Name:      plan.Name.ValueString(),
		Dimension: int(plan.Dimension.ValueInt64()),
		Metric:    metric,
		Replicas:  int(plan.Replicas.ValueInt64()),
		Pods:      int(plan.Pods.ValueInt64()),
		PodType:   podType,
	}

	// Create new order
	err = r.client.CreateIndex(ctx, item)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating index",
			"Could not create index, unexpected error: "+err.Error(),
		)
		return
	}

	var done bool
	var result *DescribeIndexResponse
	for !done {
		result, err = r.client.DescribeIndex(ctx, plan.Name.ValueString())

		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating index",
				"Could not create index, unexpected error: "+err.Error(),
			)
			return
		}
		if result != nil && result.Status.Ready {
			done = true
		}
		time.Sleep(5 * time.Second)
	}

	if result == nil {
		resp.Diagnostics.AddError(
			"Error creating index",
			"Could not create index, unexpected error: "+err.Error(),
		)
		return
	}
	// Map response body to schema and populate Computed attribute values
	plan = indexResourceModel{
		ID:        types.StringValue(result.Database.Name),
		Name:      types.StringValue(result.Database.Name),
		Dimension: types.Int64Value(int64(result.Database.Dimension)),
		Metric:    types.StringValue(result.Database.Metric.String()),
		Pods:      types.Int64Value(int64(result.Database.Pods)),
		Replicas:  types.Int64Value(int64(result.Database.Replicas)),
		PodType:   types.StringValue(result.Database.PodType.String()),
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *indexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state indexResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed index value from Pinecone
	index, err := r.client.DescribeIndex(ctx, state.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pinecone Index",
			"Could not read Pinecone Index, unexpected error: "+err.Error(),
		)
		return
	}

	// If index is nil, then the index has been deleted
	if index == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state = indexResourceModel{
		ID:        types.StringValue(index.Database.Name),
		Name:      types.StringValue(index.Database.Name),
		Dimension: types.Int64Value(int64(index.Database.Dimension)),
		Metric:    types.StringValue(index.Database.Metric.String()),
		Pods:      types.Int64Value(int64(index.Database.Pods)),
		Replicas:  types.Int64Value(int64(index.Database.Replicas)),
		PodType:   types.StringValue(index.Database.PodType.String()),
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *indexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan indexResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	podType, err := NewPodType(plan.PodType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating index",
			"Could not update index, unexpected error: "+err.Error(),
		)
		return
	}

	// Generate API request body from plan
	indexItem := ConfigureIndexRequest{
		Replicas: int(plan.Replicas.ValueInt64()),
		PodType:  podType,
	}

	err = r.client.ConfigureIndex(ctx, plan.Name.ValueString(), indexItem)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating index",
			"Could not update index, unexpected error: "+err.Error(),
		)
		return
	}
	var done bool
	var result *DescribeIndexResponse
	for !done {
		result, err = r.client.DescribeIndex(ctx, plan.Name.ValueString())

		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating index",
				"Could not update index, unexpected error: "+err.Error(),
			)
			return
		}
		if result != nil && result.Status.Ready {
			done = true
		}
		time.Sleep(5 * time.Second)
	}

	if result == nil {
		resp.Diagnostics.AddError(
			"Error updating index",
			"Could not update index, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = indexResourceModel{
		ID:        types.StringValue(result.Database.Name),
		Name:      types.StringValue(result.Database.Name),
		Dimension: types.Int64Value(int64(result.Database.Dimension)),
		Metric:    types.StringValue(result.Database.Metric.String()),
		Pods:      types.Int64Value(int64(result.Database.Pods)),
		Replicas:  types.Int64Value(int64(result.Database.Replicas)),
		PodType:   types.StringValue(result.Database.PodType.String()),
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *indexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state indexResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	err := r.client.DeleteIndex(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting index",
			"Could not delete index, unexpected error: "+err.Error(),
		)
		return
	}

	var done bool
	var result *DescribeIndexResponse
	for !done {
		result, err = r.client.DescribeIndex(ctx, state.Name.ValueString())

		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting index",
				"Could not delete index, unexpected error: "+err.Error(),
			)
			return
		}
		if result == nil {
			done = true
		}
		time.Sleep(5 * time.Second)
	}
}

// Configure adds the provider configured client to the resource.
func (r *indexResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(PineconeClientInterface)
	if !ok {
		resp.Diagnostics.AddError("Error Configure", "Invalid provider data")
		return
	}
	r.client = client
}

func (r *indexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
