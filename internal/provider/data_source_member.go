package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSource = &MemberDataSource{}
var _ datasource.DataSourceWithConfigure = &MemberDataSource{}

func NewMemberDataSource() datasource.DataSource {
	return &MemberDataSource{}
}

type MemberDataSource struct {
	client *APIClient
}

func (d *MemberDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member"
}

func (d *MemberDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Fetches a DollarBox organisation member by ID or email.",
		Attributes:          memberLookupDataSourceAttributes(),
	}
}

func memberLookupDataSourceAttributes() map[string]datasourceschema.Attribute {
	attrs := memberDataSourceAttributes()
	attrs["id"] = datasourceschema.StringAttribute{
		MarkdownDescription: "DollarBox membership ID.",
		Optional:            true,
		Computed:            true,
	}
	attrs["email"] = datasourceschema.StringAttribute{
		MarkdownDescription: "Member email address. May be used instead of `id` for lookup.",
		Optional:            true,
		Computed:            true,
	}
	return attrs
}

func memberDataSourceAttributes() map[string]datasourceschema.Attribute {
	return map[string]datasourceschema.Attribute{
		"id": datasourceschema.StringAttribute{
			MarkdownDescription: "DollarBox membership ID.",
			Computed:            true,
		},
		"email": datasourceschema.StringAttribute{
			MarkdownDescription: "Member email address.",
			Computed:            true,
		},
		"role": datasourceschema.StringAttribute{
			MarkdownDescription: "Member role in the organisation.",
			Computed:            true,
		},
		"joined_at": datasourceschema.StringAttribute{
			MarkdownDescription: "Membership join timestamp.",
			Computed:            true,
		},
	}
}

func (d *MemberDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config memberModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := ""
	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		id = strings.TrimSpace(config.ID.ValueString())
	}
	email := ""
	if !config.Email.IsNull() && !config.Email.IsUnknown() {
		email = strings.TrimSpace(config.Email.ValueString())
	}
	if id == "" && email == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Missing Member Lookup",
			"Set id or email to look up a DollarBox member.",
		)
		return
	}

	var member apiMember
	var err error
	if id != "" {
		member, err = d.client.GetMember(ctx, id)
		if err != nil {
			addAPIError(&resp.Diagnostics, "Read DollarBox Member", err)
			return
		}
		if email != "" && !strings.EqualFold(member.Email, email) {
			resp.Diagnostics.AddAttributeError(
				path.Root("email"),
				"Member Lookup Mismatch",
				fmt.Sprintf("Member %s has email %q, not %q.", id, member.Email, email),
			)
			return
		}
	} else {
		members, listErr := d.client.ListMembers(ctx)
		if listErr != nil {
			addAPIError(&resp.Diagnostics, "List DollarBox Members", listErr)
			return
		}
		var ok bool
		member, ok = findMemberByEmail(members, email)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("email"),
				"Member Not Found",
				fmt.Sprintf("No DollarBox member with email %q was found in the selected organisation.", email),
			)
			return
		}
	}

	state := memberModelFromAPI(member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
