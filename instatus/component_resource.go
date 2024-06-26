package instatus

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"

	is "github.com/brunoscota/instatus-client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &componentResource{}
	_ resource.ResourceWithConfigure   = &componentResource{}
	_ resource.ResourceWithImportState = &componentResource{}
)

// Configure adds the provider configured client to the resource.
func (r *componentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*is.Client)
}

// NewComponentResource is a helper function to simplify the provider implementation.
func NewComponentResource() resource.Resource {
	return &componentResource{}
}

// componentResource is the resource implementation.
type componentResource struct {
	client *is.Client
}

// componentResourceModel maps the resource schema data.
type componentResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PageID      types.String `tfsdk:"page_id"`
	Description types.String `tfsdk:"description"`
	ShowUptime  types.Bool   `tfsdk:"show_uptime"`
	Grouped     types.Bool   `tfsdk:"grouped"`
	GroupName   types.String `tfsdk:"group_name"`
	GroupId     types.String `tfsdk:"group_id"`
}

// Metadata returns the resource type name.
func (r *componentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

// Schema defines the schema for the resource.
func (r *componentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a component.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "String Identifier of the component.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"page_id": schema.StringAttribute{
				Description: "String Identifier of the page of the component.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the component.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the component.",
				Optional:    true,
			},
			"show_uptime": schema.BoolAttribute{
				Description: "Whether show uptime is enabled in the component.",
				Optional:    true,
			},
			"grouped": schema.BoolAttribute{
				Description: "Whether the component is in a group (Require group set to desired name when true).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"group_name": schema.StringAttribute{
				Description: "Name of the group for the component (Require grouped set to true).",
				Optional:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "Name of the group for the component (Require grouped set to true).",
				Optional:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *componentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan componentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var item is.Component = is.Component{
		Name:        plan.Name.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		ShowUptime:  plan.ShowUptime.ValueBoolPointer(),
		Grouped:     plan.Grouped.ValueBoolPointer(),
		Group:       plan.GroupId.ValueStringPointer(),
		GroupId:     plan.GroupId.ValueStringPointer(),
	}

	// Create new component
	component, err := r.client.CreateComponent(plan.PageID.ValueString(), &item)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating component",
			"Could not create component, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringPointerValue(component.ID)
	plan.Description = types.StringPointerValue(component.Description)
	plan.GroupName = types.StringPointerValue(component.Group.Name)
	plan.GroupId = types.StringPointerValue(component.Group.Id)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *componentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state componentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed component value from Instatus
	component, err := r.client.GetComponent(state.PageID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Instatus Component",
			"Could not read Instatus component ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// Overwrite items with refreshed state
	state.Name = types.StringPointerValue(component.Name)
	state.Description = types.StringPointerValue(component.Description)
	state.ShowUptime = types.BoolPointerValue(component.ShowUptime)
	state.Grouped = types.BoolValue(component.Group.Name != nil)
	state.GroupName = types.StringPointerValue(component.Group.Name)
	state.GroupId = types.StringPointerValue(component.Group.Id)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *componentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan componentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var item is.Component = is.Component{
		Name:        plan.Name.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		ShowUptime:  plan.ShowUptime.ValueBoolPointer(),
		Grouped:     plan.Grouped.ValueBoolPointer(),
		Group:       plan.GroupName.ValueStringPointer(),
		GroupId:     plan.GroupId.ValueStringPointer(),
	}

	// Update existing component
	component, err := r.client.UpdateComponent(plan.PageID.ValueString(), plan.ID.ValueString(), &item)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Instatus Component",
			"Could not update component, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringPointerValue(component.ID)
	plan.GroupName = types.StringPointerValue(component.Group.Name)
	plan.Description = types.StringPointerValue(component.Description)
	plan.GroupId = types.StringPointerValue(component.Group.Id)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *componentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state componentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing component
	err := r.client.DeleteComponent(state.PageID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Instatus Component",
			"Could not delete component, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *componentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/") // Splitting by '/' for "PageId/id"

	// Check if the split results exactly in two parts and neither part is empty
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Import identifier must be in the format 'PageId/id'. Got: "+req.ID,
		)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("page_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
