package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var _ resource.Resource = &NamespaceResource{}
var _ resource.ResourceWithConfigure = &NamespaceResource{}
var _ resource.ResourceWithImportState = &NamespaceResource{}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

type NamespaceResource struct {
	client *APIClient
}

func (r *NamespaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "A DollarBox child namespace with an allocated slice of organisation capacity.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "DollarBox namespace ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": resourceschema.StringAttribute{
				MarkdownDescription: "Namespace slug. Changing this recreates the namespace.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allocated_containers": resourceschema.Int64Attribute{
				MarkdownDescription: "Number of container slots allocated to this namespace.",
				Required:            true,
			},
			"allocated_volume_gb": resourceschema.Int64Attribute{
				MarkdownDescription: "Volume capacity in GB allocated to this namespace.",
				Required:            true,
			},
			"status": resourceschema.StringAttribute{
				MarkdownDescription: "Current namespace status.",
				Computed:            true,
			},
			"k8s_namespace": resourceschema.StringAttribute{
				MarkdownDescription: "Backing Kubernetes namespace name.",
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

func (r *NamespaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace, err := r.client.CreateNamespace(ctx, namespacePayload{
		Slug:                plan.Slug.ValueString(),
		AllocatedContainers: plan.AllocatedContainers.ValueInt64(),
		AllocatedVolumeGB:   plan.AllocatedVolumeGB.ValueInt64(),
	})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Namespace", err)
		return
	}

	updateNamespaceModelFromAPI(&plan, namespace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state namespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace, err := r.client.GetNamespace(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Namespace", err)
		return
	}

	updateNamespaceModelFromAPI(&state, namespace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan namespaceModel
	var state namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace, err := r.client.UpdateNamespace(ctx, state.ID.ValueString(), namespacePayload{
		AllocatedContainers: plan.AllocatedContainers.ValueInt64(),
		AllocatedVolumeGB:   plan.AllocatedVolumeGB.ValueInt64(),
	})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Update DollarBox Namespace", err)
		return
	}

	updateNamespaceModelFromAPI(&plan, namespace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state namespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNamespace(ctx, state.ID.ValueString())
	if err != nil && !isNotFoundError(err) {
		addAPIError(&resp.Diagnostics, "Delete DollarBox Namespace", err)
	}
}

func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
