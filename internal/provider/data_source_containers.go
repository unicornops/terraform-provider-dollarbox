package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ContainersDataSource{}
var _ datasource.DataSourceWithConfigure = &ContainersDataSource{}

func NewContainersDataSource() datasource.DataSource {
	return &ContainersDataSource{}
}

type ContainersDataSource struct {
	client *APIClient
}

type containersDataSourceModel struct {
	Containers []containerResourceModel `tfsdk:"containers"`
}

func (d *ContainersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers"
}

func (d *ContainersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists DollarBox containers.",
		Attributes: map[string]datasourceschema.Attribute{
			"containers": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Containers in the selected organisation.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
					"id":           datasourceschema.StringAttribute{MarkdownDescription: "DollarBox container ID.", Computed: true},
					"name":         datasourceschema.StringAttribute{MarkdownDescription: "Container name.", Computed: true},
					"image":        datasourceschema.StringAttribute{MarkdownDescription: "OCI image reference.", Computed: true},
					"env":          datasourceschema.MapAttribute{MarkdownDescription: "Environment variables for the container.", ElementType: types.StringType, Computed: true, Sensitive: true},
					"command":      datasourceschema.ListAttribute{MarkdownDescription: "Container command override.", ElementType: types.StringType, Computed: true},
					"status":       datasourceschema.StringAttribute{MarkdownDescription: "Current container status.", Computed: true},
					"ipv6_address": datasourceschema.StringAttribute{MarkdownDescription: "Assigned IPv6 LoadBalancer address.", Computed: true},
					"created_at":   datasourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
					"updated_at":   datasourceschema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
				}},
			},
		},
	}
}

func (d *ContainersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ContainersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	containers, err := d.client.ListContainers(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Containers", err)
		return
	}
	state := containersDataSourceModel{Containers: make([]containerResourceModel, 0, len(containers))}
	for _, container := range containers {
		model := containerResourceModel{}
		resp.Diagnostics.Append(model.updateFromContainer(ctx, container)...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Containers = append(state.Containers, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
