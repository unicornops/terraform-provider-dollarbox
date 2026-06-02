package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &InvitationResource{}
var _ resource.ResourceWithConfigure = &InvitationResource{}
var _ resource.ResourceWithImportState = &InvitationResource{}

func NewInvitationResource() resource.Resource {
	return &InvitationResource{}
}

type InvitationResource struct {
	client *APIClient
}

type invitationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	Accepted  types.Bool   `tfsdk:"accepted"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *InvitationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invitation"
}

func (r *InvitationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "A DollarBox organisation invitation.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "DollarBox invitation ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": resourceschema.StringAttribute{
				MarkdownDescription: "Email address to invite. Changing this recreates the invitation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": resourceschema.StringAttribute{
				MarkdownDescription: "Role to grant when the invitation is accepted. Defaults to `member`. The API rejects `owner` invitations.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"accepted": resourceschema.BoolAttribute{
				MarkdownDescription: "Whether the invitation has been accepted.",
				Computed:            true,
			},
			"expires_at": resourceschema.StringAttribute{
				MarkdownDescription: "Invitation expiry timestamp.",
				Computed:            true,
			},
			"created_at": resourceschema.StringAttribute{
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *InvitationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InvitationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan invitationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := invitationPayload{Email: plan.Email.ValueString()}
	if !plan.Role.IsNull() && !plan.Role.IsUnknown() {
		payload.Role = plan.Role.ValueString()
	}
	invitation, err := r.client.CreateInvitation(ctx, payload)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Invitation", err)
		return
	}

	plan.updateFromInvitation(invitation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *InvitationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state invitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invitation, err := r.client.GetInvitation(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Invitation", err)
		return
	}

	state.updateFromInvitation(invitation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *InvitationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected DollarBox Invitation Update",
		"DollarBox invitations cannot be updated through Terraform. Terraform should plan a replacement when email or role changes.",
	)
}

func (r *InvitationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state invitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteInvitation(ctx, state.ID.ValueString())
	if err != nil && !isNotFoundError(err) {
		addAPIError(&resp.Diagnostics, "Delete DollarBox Invitation", err)
	}
}

func (r *InvitationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *invitationResourceModel) updateFromInvitation(invitation apiInvitation) {
	m.ID = types.StringValue(strconv.FormatInt(invitation.ID, 10))
	m.Email = types.StringValue(invitation.Email)
	m.Role = types.StringValue(invitation.Role)
	m.Accepted = types.BoolValue(invitation.Accepted)
	m.ExpiresAt = types.StringValue(invitation.ExpiresAt)
	m.CreatedAt = types.StringValue(invitation.CreatedAt)
}
