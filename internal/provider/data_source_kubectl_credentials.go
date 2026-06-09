package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &KubectlCredentialsDataSource{}
var _ datasource.DataSourceWithConfigure = &KubectlCredentialsDataSource{}

func NewKubectlCredentialsDataSource() datasource.DataSource { return &KubectlCredentialsDataSource{} }

type KubectlCredentialsDataSource struct{ client *APIClient }

type kubectlCredentialsDataSourceModel struct {
	KubectlCredentials []kubectlCredentialResourceModel `tfsdk:"kubectl_credentials"`
}

func (d *KubectlCredentialsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubectl_credentials"
}

func (d *KubectlCredentialsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists DollarBox kubectl credentials.",
		Attributes: map[string]datasourceschema.Attribute{
			"kubectl_credentials": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Kubectl credentials for the authenticated user.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
					"id":                 datasourceschema.StringAttribute{MarkdownDescription: "DollarBox kubectl credential ID.", Computed: true},
					"org":                datasourceschema.StringAttribute{MarkdownDescription: "Organisation slug for the credential.", Computed: true},
					"sa_name":            datasourceschema.StringAttribute{MarkdownDescription: "Kubernetes service account name.", Computed: true},
					"kubeconfig":         datasourceschema.StringAttribute{MarkdownDescription: "Kubeconfig returned only when a credential is issued.", Computed: true, Sensitive: true},
					"created_at":         datasourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
					"rotated_at":         datasourceschema.StringAttribute{MarkdownDescription: "Last rotation timestamp.", Computed: true},
					"last_downloaded_at": datasourceschema.StringAttribute{MarkdownDescription: "Last kubeconfig download timestamp.", Computed: true},
				}},
			},
		},
	}
}

func (d *KubectlCredentialsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KubectlCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	credentials, err := d.client.ListKubectlCredentials(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Kubectl Credentials", err)
		return
	}
	state := kubectlCredentialsDataSourceModel{KubectlCredentials: make([]kubectlCredentialResourceModel, 0, len(credentials))}
	for _, credential := range credentials {
		model := kubectlCredentialResourceModel{}
		model.updateFromKubectlCredential(credential, types.StringNull())
		state.KubectlCredentials = append(state.KubectlCredentials, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
