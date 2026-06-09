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

var _ resource.Resource = &MemberResource{}
var _ resource.ResourceWithConfigure = &MemberResource{}
var _ resource.ResourceWithImportState = &MemberResource{}

func NewMemberResource() resource.Resource {
	return &MemberResource{}
}

type MemberResource struct {
	client *APIClient
}

func (r *MemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member"
}

func (r *MemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "A DollarBox organisation member. The API does not create members directly; this resource adopts an existing member by email, manages role, and removes the membership on destroy.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "DollarBox membership ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": resourceschema.StringAttribute{
				MarkdownDescription: "Email address of an existing organisation member to adopt. Changing this adopts a different member and removes the old membership during replacement.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": resourceschema.StringAttribute{
				MarkdownDescription: "Member role in the organisation.",
				Required:            true,
			},
			"joined_at": resourceschema.StringAttribute{
				MarkdownDescription: "Membership join timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *MemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	members, err := r.client.ListMembers(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Members", err)
		return
	}
	member, ok := findMemberByEmail(members, plan.Email.ValueString())
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"Member Not Found",
			"DollarBox does not create members directly through /api/v1/members/. Invite the user first, wait for acceptance, then manage the accepted member.",
		)
		return
	}
	if member.ID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"Member Missing ID",
			"The DollarBox API returned a matching member without an id, so Terraform cannot manage it.",
		)
		return
	}

	if member.Role != plan.Role.ValueString() {
		member, err = r.client.UpdateMember(ctx, member.ID, memberPayload{Role: plan.Role.ValueString()})
		if err != nil {
			addAPIError(&resp.Diagnostics, "Update DollarBox Member", err)
			return
		}
	}

	updateMemberModelFromAPI(&plan, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state memberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GetMember(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Member", err)
		return
	}

	updateMemberModelFromAPI(&state, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan memberModel
	var state memberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.UpdateMember(ctx, state.ID.ValueString(), memberPayload{Role: plan.Role.ValueString()})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Update DollarBox Member", err)
		return
	}

	updateMemberModelFromAPI(&plan, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state memberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMember(ctx, state.ID.ValueString())
	if err != nil && !isNotFoundError(err) {
		addAPIError(&resp.Diagnostics, "Delete DollarBox Member", err)
	}
}

func (r *MemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
