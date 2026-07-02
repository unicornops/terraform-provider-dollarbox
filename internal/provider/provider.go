package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultEndpoint = "https://app.dollarbox.dev"

var _ provider.Provider = &DollarBoxProvider{}

type DollarBoxProvider struct {
	version string
}

type DollarBoxProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
	Org      types.String `tfsdk:"org"`
}

type ClientConfig struct {
	Endpoint string
	Token    string
	Org      string
}

func (p *DollarBoxProvider) Metadata(
	ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "dollarbox"
	resp.Version = p.version
}

func (p *DollarBoxProvider) Schema(
	ctx context.Context,
	req provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The DollarBox provider manages DollarBox resources through the public API.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "DollarBox API endpoint. Defaults to `https://app.dollarbox.dev`.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "DollarBox API token. May also be set with `DOLLARBOX_TOKEN`.",
				Optional:            true,
				Sensitive:           true,
			},
			"org": schema.StringAttribute{
				MarkdownDescription: "Optional DollarBox organisation slug used as the default scope.",
				Optional:            true,
			},
		},
	}
}

func (p *DollarBoxProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data DollarBoxProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Endpoint.IsUnknown() || data.Token.IsUnknown() || data.Org.IsUnknown() {
		return
	}

	config := &ClientConfig{Endpoint: defaultEndpoint}

	if !data.Endpoint.IsNull() && !data.Endpoint.IsUnknown() {
		endpoint := strings.TrimRight(strings.TrimSpace(data.Endpoint.ValueString()), "/")
		if endpoint == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("endpoint"),
				"Invalid Endpoint",
				"Endpoint must not be empty when it is configured.",
			)
			return
		}
		config.Endpoint = endpoint
	}

	if !data.Token.IsNull() && !data.Token.IsUnknown() {
		token := strings.TrimSpace(data.Token.ValueString())
		if token == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("token"),
				"Invalid Token",
				"Token must not be empty when it is configured.",
			)
			return
		}
		config.Token = token
	} else if token := strings.TrimSpace(os.Getenv("DOLLARBOX_TOKEN")); token != "" {
		config.Token = token
	}
	if config.Token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Token",
			"Configure a DollarBox API token with the token attribute or the DOLLARBOX_TOKEN environment variable.",
		)
		return
	}

	if !data.Org.IsNull() && !data.Org.IsUnknown() {
		org := strings.TrimSpace(data.Org.ValueString())
		if org == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("org"),
				"Invalid Org",
				"Org must not be empty when it is configured.",
			)
			return
		}
		config.Org = org
	}

	client := NewAPIClient(*config)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DollarBoxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewContainerResource,
		NewVolumeResource,
		NewInvitationResource,
		NewKubectlCredentialResource,
		NewMemberResource,
		NewNamespaceResource,
		NewOrgResource,
		NewSnapshotPolicyResource,
		NewVolumeSnapshotResource,
		NewSnapshotRestoreResource,
	}
}

func (p *DollarBoxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrgDataSource,
		NewContainersDataSource,
		NewVolumesDataSource,
		NewInvitationsDataSource,
		NewKubectlCredentialsDataSource,
		NewMemberDataSource,
		NewMembersDataSource,
		NewNamespaceDataSource,
		NewNamespacesDataSource,
		NewOrgsDataSource,
		NewVolumeSnapshotDataSource,
		NewVolumeSnapshotsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DollarBoxProvider{version: version}
	}
}
