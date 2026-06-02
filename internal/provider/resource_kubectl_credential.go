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

var _ resource.Resource = &KubectlCredentialResource{}
var _ resource.ResourceWithConfigure = &KubectlCredentialResource{}
var _ resource.ResourceWithImportState = &KubectlCredentialResource{}

func NewKubectlCredentialResource() resource.Resource {
	return &KubectlCredentialResource{}
}

type KubectlCredentialResource struct {
	client *APIClient
}

type kubectlCredentialResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Org              types.String `tfsdk:"org"`
	SAName           types.String `tfsdk:"sa_name"`
	Kubeconfig       types.String `tfsdk:"kubeconfig"`
	CreatedAt        types.String `tfsdk:"created_at"`
	RotatedAt        types.String `tfsdk:"rotated_at"`
	LastDownloadedAt types.String `tfsdk:"last_downloaded_at"`
}

func (r *KubectlCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubectl_credential"
}

func (r *KubectlCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "A DollarBox kubectl credential for the authenticated user and selected organisation.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "DollarBox kubectl credential ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org": resourceschema.StringAttribute{
				MarkdownDescription: "Organisation slug for the credential.",
				Computed:            true,
			},
			"sa_name": resourceschema.StringAttribute{
				MarkdownDescription: "Kubernetes service account name.",
				Computed:            true,
			},
			"kubeconfig": resourceschema.StringAttribute{
				MarkdownDescription: "Kubeconfig returned when the credential is issued. The API only returns this value on create.",
				Computed:            true,
				Sensitive:           true,
			},
			"created_at": resourceschema.StringAttribute{
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
			"rotated_at": resourceschema.StringAttribute{
				MarkdownDescription: "Last rotation timestamp.",
				Computed:            true,
			},
			"last_downloaded_at": resourceschema.StringAttribute{
				MarkdownDescription: "Last kubeconfig download timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *KubectlCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KubectlCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kubectlCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credential, err := r.client.CreateKubectlCredential(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Create DollarBox Kubectl Credential", err)
		return
	}

	plan.updateFromKubectlCredential(credential, types.StringNull())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KubectlCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state kubectlCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credential, err := r.client.GetKubectlCredential(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addAPIError(&resp.Diagnostics, "Read DollarBox Kubectl Credential", err)
		return
	}

	state.updateFromKubectlCredential(credential, state.Kubeconfig)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KubectlCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected DollarBox Kubectl Credential Update",
		"DollarBox kubectl credentials have no configurable Terraform attributes.",
	)
}

func (r *KubectlCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kubectlCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteKubectlCredential(ctx, state.ID.ValueString())
	if err != nil && !isNotFoundError(err) {
		addAPIError(&resp.Diagnostics, "Delete DollarBox Kubectl Credential", err)
	}
}

func (r *KubectlCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *kubectlCredentialResourceModel) updateFromKubectlCredential(credential apiKubectlCredential, preservedKubeconfig types.String) {
	m.ID = types.StringValue(strconv.FormatInt(credential.ID, 10))
	m.Org = types.StringValue(credential.Org)
	m.SAName = types.StringValue(credential.SAName)
	m.CreatedAt = types.StringValue(credential.CreatedAt)
	m.RotatedAt = types.StringValue(credential.RotatedAt)
	if credential.LastDownloadedAt == nil {
		m.LastDownloadedAt = types.StringNull()
	} else {
		m.LastDownloadedAt = types.StringValue(*credential.LastDownloadedAt)
	}
	if credential.Kubeconfig != "" {
		m.Kubeconfig = types.StringValue(credential.Kubeconfig)
	} else {
		m.Kubeconfig = preservedKubeconfig
	}
}
