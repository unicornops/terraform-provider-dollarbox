package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &InvitationsDataSource{}
var _ datasource.DataSourceWithConfigure = &InvitationsDataSource{}

func NewInvitationsDataSource() datasource.DataSource { return &InvitationsDataSource{} }

type InvitationsDataSource struct{ client *APIClient }

type invitationsDataSourceModel struct {
	Invitations []invitationResourceModel `tfsdk:"invitations"`
}

func (d *InvitationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invitations"
}

func (d *InvitationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists DollarBox organisation invitations.",
		Attributes: map[string]datasourceschema.Attribute{
			"invitations": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Invitations in the selected organisation.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
					"id":         datasourceschema.StringAttribute{MarkdownDescription: "DollarBox invitation ID.", Computed: true},
					"email":      datasourceschema.StringAttribute{MarkdownDescription: "Invited email address.", Computed: true},
					"role":       datasourceschema.StringAttribute{MarkdownDescription: "Role granted when accepted.", Computed: true},
					"accepted":   datasourceschema.BoolAttribute{MarkdownDescription: "Whether the invitation has been accepted.", Computed: true},
					"expires_at": datasourceschema.StringAttribute{MarkdownDescription: "Invitation expiry timestamp.", Computed: true},
					"created_at": datasourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
				}},
			},
		},
	}
}

func (d *InvitationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *APIClient, got %T. Please report this to the provider developers.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *InvitationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	invitations, err := d.client.ListInvitations(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Invitations", err)
		return
	}
	state := invitationsDataSourceModel{Invitations: make([]invitationResourceModel, 0, len(invitations))}
	for _, invitation := range invitations {
		model := invitationResourceModel{}
		model.updateFromInvitation(invitation)
		state.Invitations = append(state.Invitations, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
