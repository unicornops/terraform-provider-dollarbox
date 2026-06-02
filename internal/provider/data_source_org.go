package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OrgDataSource{}
var _ datasource.DataSourceWithConfigure = &OrgDataSource{}

func NewOrgDataSource() datasource.DataSource {
	return &OrgDataSource{}
}

type OrgDataSource struct {
	client *APIClient
}

type orgDataSourceModel struct {
	Slug           types.String `tfsdk:"slug"`
	Name           types.String `tfsdk:"name"`
	BillingEmail   types.String `tfsdk:"billing_email"`
	Status         types.String `tfsdk:"status"`
	BillingMode    types.String `tfsdk:"billing_mode"`
	KubectlEnabled types.Bool   `tfsdk:"kubectl_enabled"`
	APIEnabled     types.Bool   `tfsdk:"api_enabled"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *OrgDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *OrgDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Fetches a DollarBox organisation by slug.",
		Attributes: map[string]datasourceschema.Attribute{
			"slug": datasourceschema.StringAttribute{
				MarkdownDescription: "Organisation slug. Defaults to the provider-level org when omitted.",
				Optional:            true,
				Computed:            true,
			},
			"name": datasourceschema.StringAttribute{
				MarkdownDescription: "Organisation display name.",
				Computed:            true,
			},
			"billing_email": datasourceschema.StringAttribute{
				MarkdownDescription: "Billing email address.",
				Computed:            true,
			},
			"status": datasourceschema.StringAttribute{
				MarkdownDescription: "Organisation lifecycle status.",
				Computed:            true,
			},
			"billing_mode": datasourceschema.StringAttribute{
				MarkdownDescription: "Organisation billing mode.",
				Computed:            true,
			},
			"kubectl_enabled": datasourceschema.BoolAttribute{
				MarkdownDescription: "Whether kubectl mode is enabled for this organisation.",
				Computed:            true,
			},
			"api_enabled": datasourceschema.BoolAttribute{
				MarkdownDescription: "Whether API access is enabled for this organisation.",
				Computed:            true,
			},
			"created_at": datasourceschema.StringAttribute{
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
		},
	}
}

func (d *OrgDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *APIClient, got %T. Please report this to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *OrgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config orgDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Provider Not Configured", "The DollarBox provider client was not configured.")
		return
	}

	slug := ""
	if !config.Slug.IsNull() && !config.Slug.IsUnknown() {
		slug = strings.TrimSpace(config.Slug.ValueString())
	}
	if slug == "" {
		slug = d.client.org
	}
	if slug == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("slug"),
			"Missing Organisation Slug",
			"Set slug on the data source or configure org on the provider.",
		)
		return
	}

	org, err := d.client.GetOrg(ctx, slug)
	if err != nil {
		addAPIError(&resp.Diagnostics, "Read DollarBox Org", err)
		return
	}

	state := orgDataSourceModel{
		Slug:           types.StringValue(org.Slug),
		Name:           types.StringValue(org.Name),
		BillingEmail:   types.StringValue(org.BillingEmail),
		Status:         types.StringValue(org.Status),
		BillingMode:    types.StringValue(org.BillingMode),
		KubectlEnabled: types.BoolValue(org.KubectlEnabled),
		APIEnabled:     types.BoolValue(org.APIEnabled),
		CreatedAt:      types.StringValue(org.CreatedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
