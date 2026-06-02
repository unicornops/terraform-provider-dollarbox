package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ContainerResource{}
var _ resource.ResourceWithConfigure = &ContainerResource{}
var _ resource.ResourceWithImportState = &ContainerResource{}

func NewContainerResource() resource.Resource {
	return &ContainerResource{}
}

type ContainerResource struct {
	client *APIClient
}

type containerResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Image       types.String `tfsdk:"image"`
	Env         types.Map    `tfsdk:"env"`
	Command     types.List   `tfsdk:"command"`
	Status      types.String `tfsdk:"status"`
	IPv6Address types.String `tfsdk:"ipv6_address"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *ContainerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

func (r *ContainerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "A DollarBox container.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "DollarBox container ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": resourceschema.StringAttribute{
				MarkdownDescription: "Container name. Must be unique in the selected org. Changing the name recreates the container.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": resourceschema.StringAttribute{
				MarkdownDescription: "OCI image reference to run.",
				Required:            true,
			},
			"env": resourceschema.MapAttribute{
				MarkdownDescription: "Environment variables for the container.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"command": resourceschema.ListAttribute{
				MarkdownDescription: "Container command override.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"status": resourceschema.StringAttribute{
				MarkdownDescription: "Current container status.",
				Computed:            true,
			},
			"ipv6_address": resourceschema.StringAttribute{
				MarkdownDescription: "Assigned IPv6 LoadBalancer address.",
				Computed:            true,
			},
			"created_at": resourceschema.StringAttribute{
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
			"updated_at": resourceschema.StringAttribute{
				MarkdownDescription: "Last update timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *ContainerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *APIClient, got %T. Please report this to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *ContainerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan containerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.toContainerPayload(ctx, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	container, err := r.client.CreateContainer(ctx, payload)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Container", err)
		return
	}

	resp.Diagnostics.Append(plan.updateFromContainer(ctx, container)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContainerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	container, err := r.client.GetContainer(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Container", err)
		return
	}

	resp.Diagnostics.Append(state.updateFromContainer(ctx, container)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContainerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containerResourceModel
	var state containerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.toContainerPayload(ctx, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	container, err := r.client.UpdateContainer(ctx, state.ID.ValueString(), payload)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Update DollarBox Container", err)
		return
	}

	resp.Diagnostics.Append(plan.updateFromContainer(ctx, container)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContainerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state containerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteContainer(ctx, state.ID.ValueString())
	if err != nil && !isNotFoundError(err) {
		addAPIError(&resp.Diagnostics, "Delete DollarBox Container", err)
	}
}

func (r *ContainerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m containerResourceModel) toContainerPayload(ctx context.Context, includeName bool) (containerPayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	env, envDiags := expandStringMap(ctx, m.Env)
	command, commandDiags := expandStringList(ctx, m.Command)
	diags.Append(envDiags...)
	diags.Append(commandDiags...)

	payload := containerPayload{
		Image:   m.Image.ValueString(),
		Env:     env,
		Command: command,
	}
	if includeName {
		payload.Name = m.Name.ValueString()
	}
	return payload, diags
}

func (m *containerResourceModel) updateFromContainer(ctx context.Context, container apiContainer) diag.Diagnostics {
	var diags diag.Diagnostics
	m.ID = types.StringValue(strconv.FormatInt(container.ID, 10))
	m.Name = types.StringValue(container.Name)
	m.Image = types.StringValue(container.Image)
	m.Status = types.StringValue(container.Status)
	m.IPv6Address = types.StringValue(container.IPv6Address)
	m.CreatedAt = types.StringValue(container.CreatedAt)
	m.UpdatedAt = types.StringValue(container.UpdatedAt)

	if container.Env == nil {
		container.Env = map[string]string{}
	}
	if container.Command == nil {
		container.Command = []string{}
	}
	env, envDiags := types.MapValueFrom(ctx, types.StringType, container.Env)
	command, commandDiags := types.ListValueFrom(ctx, types.StringType, container.Command)
	diags.Append(envDiags...)
	diags.Append(commandDiags...)
	m.Env = env
	m.Command = command
	return diags
}

func expandStringMap(ctx context.Context, value types.Map) (map[string]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return map[string]string{}, nil
	}
	result := map[string]string{}
	diags := value.ElementsAs(ctx, &result, false)
	return result, diags
}

func expandStringList(ctx context.Context, value types.List) ([]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return []string{}, nil
	}
	result := []string{}
	diags := value.ElementsAs(ctx, &result, false)
	return result, diags
}
