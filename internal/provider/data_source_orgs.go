package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OrgsDataSource{}
var _ datasource.DataSourceWithConfigure = &OrgsDataSource{}

func NewOrgsDataSource() datasource.DataSource {
	return &OrgsDataSource{}
}

type OrgsDataSource struct {
	client *APIClient
}

type orgsDataSourceModel struct {
	Orgs []orgDataSourceModel `tfsdk:"orgs"`
}

func (d *OrgsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_orgs"
}

func (d *OrgsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists DollarBox organisations the token's user belongs to.",
		Attributes: map[string]datasourceschema.Attribute{
			"orgs": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Organisations visible to the token.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: orgDataSourceAttributes(),
				},
			},
		},
	}
}

func orgDataSourceAttributes() map[string]datasourceschema.Attribute {
	return map[string]datasourceschema.Attribute{
		"slug": datasourceschema.StringAttribute{
			MarkdownDescription: "Organisation slug.",
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
	}
}

func orgDataSourceModelFromAPI(org apiOrg) orgDataSourceModel {
	return orgDataSourceModel{
		Slug:           types.StringValue(org.Slug),
		Name:           types.StringValue(org.Name),
		BillingEmail:   types.StringValue(org.BillingEmail),
		Status:         types.StringValue(org.Status),
		BillingMode:    types.StringValue(org.BillingMode),
		KubectlEnabled: types.BoolValue(org.KubectlEnabled),
		APIEnabled:     types.BoolValue(org.APIEnabled),
		CreatedAt:      types.StringValue(org.CreatedAt),
	}
}

func (d *OrgsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrgsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	orgs, err := d.client.ListOrgs(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Orgs", err)
		return
	}

	state := orgsDataSourceModel{Orgs: make([]orgDataSourceModel, 0, len(orgs))}
	for _, org := range orgs {
		state.Orgs = append(state.Orgs, orgDataSourceModelFromAPI(org))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
