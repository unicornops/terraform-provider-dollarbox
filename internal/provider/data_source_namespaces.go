package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &NamespacesDataSource{}
var _ datasource.DataSourceWithConfigure = &NamespacesDataSource{}

func NewNamespacesDataSource() datasource.DataSource {
	return &NamespacesDataSource{}
}

type NamespacesDataSource struct {
	client *APIClient
}

type namespacesDataSourceModel struct {
	Namespaces []namespaceModel `tfsdk:"namespaces"`
}

func (d *NamespacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespaces"
}

func (d *NamespacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists DollarBox namespaces.",
		Attributes: map[string]datasourceschema.Attribute{
			"namespaces": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Namespaces in the selected organisation.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: namespaceDataSourceAttributes(),
				},
			},
		},
	}
}

func (d *NamespacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NamespacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	namespaces, err := d.client.ListNamespaces(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Namespaces", err)
		return
	}

	state := namespacesDataSourceModel{Namespaces: make([]namespaceModel, 0, len(namespaces))}
	for _, namespace := range namespaces {
		state.Namespaces = append(state.Namespaces, namespaceModelFromAPI(namespace))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
