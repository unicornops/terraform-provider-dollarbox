package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &VolumesDataSource{}
var _ datasource.DataSourceWithConfigure = &VolumesDataSource{}

func NewVolumesDataSource() datasource.DataSource {
	return &VolumesDataSource{}
}

type VolumesDataSource struct {
	client *APIClient
}

type volumesDataSourceModel struct {
	Volumes []volumeResourceModel `tfsdk:"volumes"`
}

func (d *VolumesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

func (d *VolumesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists DollarBox volumes.",
		Attributes: map[string]datasourceschema.Attribute{
			"volumes": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Volumes in the selected organisation.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
					"id":            datasourceschema.StringAttribute{MarkdownDescription: "DollarBox volume ID.", Computed: true},
					"name":          datasourceschema.StringAttribute{MarkdownDescription: "Volume name.", Computed: true},
					"size_gb":       datasourceschema.Int64Attribute{MarkdownDescription: "Volume size in GB.", Computed: true},
					"status":        datasourceschema.StringAttribute{MarkdownDescription: "Current volume status.", Computed: true},
					"storage_class": datasourceschema.StringAttribute{MarkdownDescription: "Kubernetes storage class backing the volume.", Computed: true},
					"created_at":    datasourceschema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
					"updated_at":    datasourceschema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
				}},
			},
		},
	}
}

func (d *VolumesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VolumesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	volumes, err := d.client.ListVolumes(ctx)
	if err != nil {
		addAPIError(&resp.Diagnostics, "List DollarBox Volumes", err)
		return
	}
	state := volumesDataSourceModel{Volumes: make([]volumeResourceModel, 0, len(volumes))}
	for _, volume := range volumes {
		model := volumeResourceModel{}
		model.updateFromVolume(volume)
		state.Volumes = append(state.Volumes, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
