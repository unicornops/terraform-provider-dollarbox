package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &MembersDataSource{}
var _ datasource.DataSourceWithConfigure = &MembersDataSource{}

func NewMembersDataSource() datasource.DataSource {
	return &MembersDataSource{}
}

type MembersDataSource struct {
	client *APIClient
}

type membersDataSourceModel struct {
	Members []memberModel `tfsdk:"members"`
}

func (d *MembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_members"
}

func (d *MembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists DollarBox organisation members.",
		Attributes: map[string]datasourceschema.Attribute{
			"members": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Members in the selected organisation.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: memberDataSourceAttributes(),
				},
			},
		},
	}
}

func (d *MembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	members, err := d.client.ListMembers(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Members", err)
		return
	}

	state := membersDataSourceModel{Members: make([]memberModel, 0, len(members))}
	for _, member := range members {
		state.Members = append(state.Members, memberModelFromAPI(member))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
