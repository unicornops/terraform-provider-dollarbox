package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &OrgResource{}
var _ resource.ResourceWithConfigure = &OrgResource{}
var _ resource.ResourceWithImportState = &OrgResource{}

func NewOrgResource() resource.Resource {
	return &OrgResource{}
}

type OrgResource struct {
	client *APIClient
}

type orgResourceModel struct {
	Slug           types.String `tfsdk:"slug"`
	Name           types.String `tfsdk:"name"`
	BillingEmail   types.String `tfsdk:"billing_email"`
	Status         types.String `tfsdk:"status"`
	BillingMode    types.String `tfsdk:"billing_mode"`
	KubectlEnabled types.Bool   `tfsdk:"kubectl_enabled"`
	APIEnabled     types.Bool   `tfsdk:"api_enabled"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (r *OrgResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (r *OrgResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "DollarBox organisation settings. The API does not create or delete orgs through Terraform; this resource adopts an existing org by slug and manages its name and billing email.",
		Attributes: map[string]resourceschema.Attribute{
			"slug": resourceschema.StringAttribute{
				MarkdownDescription: "Organisation slug. Changing this adopts a different org and removes the old one from Terraform state during replacement.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": resourceschema.StringAttribute{
				MarkdownDescription: "Organisation display name.",
				Required:            true,
			},
			"billing_email": resourceschema.StringAttribute{
				MarkdownDescription: "Billing email address.",
				Required:            true,
			},
			"status": resourceschema.StringAttribute{
				MarkdownDescription: "Organisation lifecycle status.",
				Computed:            true,
			},
			"billing_mode": resourceschema.StringAttribute{
				MarkdownDescription: "Organisation billing mode.",
				Computed:            true,
			},
			"kubectl_enabled": resourceschema.BoolAttribute{
				MarkdownDescription: "Whether kubectl mode is enabled for this organisation.",
				Computed:            true,
			},
			"api_enabled": resourceschema.BoolAttribute{
				MarkdownDescription: "Whether API access is enabled for this organisation.",
				Computed:            true,
			},
			"created_at": resourceschema.StringAttribute{
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *OrgResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan orgResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.UpdateOrg(ctx, plan.Slug.ValueString(), orgPayload{
		Name:         plan.Name.ValueString(),
		BillingEmail: plan.BillingEmail.ValueString(),
	})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Update DollarBox Org", err)
		return
	}

	plan.updateFromOrg(org)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state orgResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrg(ctx, state.Slug.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Org", err)
		return
	}

	state.updateFromOrg(org)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OrgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan orgResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.UpdateOrg(ctx, plan.Slug.ValueString(), orgPayload{
		Name:         plan.Name.ValueString(),
		BillingEmail: plan.BillingEmail.ValueString(),
	})
	if err != nil {
		addAPIError(&resp.Diagnostics, "Update DollarBox Org", err)
		return
	}

	plan.updateFromOrg(org)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// The DollarBox API has no org delete endpoint. Destroy removes Terraform management only.
}

func (r *OrgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

func (m *orgResourceModel) updateFromOrg(org apiOrg) {
	m.Slug = types.StringValue(org.Slug)
	m.Name = types.StringValue(org.Name)
	m.BillingEmail = types.StringValue(org.BillingEmail)
	m.Status = types.StringValue(org.Status)
	m.BillingMode = types.StringValue(org.BillingMode)
	m.KubectlEnabled = types.BoolValue(org.KubectlEnabled)
	m.APIEnabled = types.BoolValue(org.APIEnabled)
	m.CreatedAt = types.StringValue(org.CreatedAt)
}
